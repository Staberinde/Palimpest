package main

import (
    "errors"
    "io"
    "io/ioutil"
    "os"
    "strconv"
    "strings"
    "time"

    "github.com/golang/glog"
    "github.com/jinzhu/gorm"
    _ "github.com/jinzhu/gorm/dialects/postgres"
    "golang.org/x/net/html"
)

type BaseModel struct {
    gorm.Model
}

type Note struct {
    BaseModel
    ExternalID                string `gorm:"unique"`
    Content                   string `gorm:"not null"`
    Source                    string `gorm:"not null"`
    Tags                      []Tag  `gorm:"many2many:note_tag_mapping;AssociationForeignKey:ID;ForeignKey:ID;"`
    OriginalCreationTimestamp time.Time
}

type Tag struct {
    BaseModel
    Name  string `gorm:"not null;unique"`
    Notes []Note `gorm:"many2many:note_tag_mapping;not null;AssociationForeignKey:ID;ForeignKey:ID;"`
}

func contains(s []string, e string) bool {
    for _, a := range s {
        if a == e {
            return true
        }
    }
    return false
}

func main() {
    glog.Info("Starting Palimpest!")
    var absPath string
    if len(os.Args) > 1 {
        absPath = os.Args[1]
    } else {
        absPath = os.Getenv("CATCH_DIRECTORY")
    }
    db := setupDatabase(
        os.Getenv("DATABASE_NAME"),
        os.Getenv("DATABASE_HOST"),
        os.Getenv("DATABASE_USER"),
        os.Getenv("DATABASE_PASSWORD"),
    )
    defer db.Close()
    openDataAndIngest(db, absPath)
    glog.Info("Expected termination of Palimpest!")
}

func openDataAndIngest(db *gorm.DB, absPath string) {
    notes := openAndProcessData(absPath)
    ingestData(notes, db)
}

func setupDatabase(
    databaseName string,
    databaseHost string,
    databaseUser string,
    databasePassword string,
) *gorm.DB {
    gormDatabaseParams := (
        "dbname=" + databaseName +
        " host=" + databaseHost +
        " user=" + databaseUser +
        " sslmode=disable " +
        " password=" + databasePassword)
    glog.Info(
        "Instantiating database with: " + gormDatabaseParams,
    )
    db, err := gorm.Open(
        "postgres",
        gormDatabaseParams,
    )
    if err != nil {
        db.Close()
        panic(err)
    }
    tx := db.Begin()
    defer tx.Commit()
    db.Exec(`
        CREATE TABLE IF NOT EXISTS notes (
            id serial PRIMARY KEY,
            external_id text UNIQUE,
            content text NOT NULL,
            source text NOT NULL,
            original_creation_timestamp timestamp with time zone NOT NULL,
            created_at timestamp with time zone,
            updated_at timestamp with time zone,
            deleted_at timestamp with time zone
        );

        CREATE TABLE IF NOT EXISTS tags (
            id serial PRIMARY KEY,
            name text NOT NULL UNIQUE,
            created_at timestamp with time zone,
            updated_at timestamp with time zone,
            deleted_at timestamp with time zone
        );

        CREATE TABLE IF NOT EXISTS note_tag_mapping (
            id serial PRIMARY KEY,
            tag_id integer NOT NULL,
            note_id integer NOT NULL,
            FOREIGN KEY (tag_id) REFERENCES tags (id),
            FOREIGN KEY (note_id) REFERENCES notes (id)
        );
    `)
    return db
}

func openAndProcessData(absPath string) []Note {
    directories, err := ioutil.ReadDir(absPath)
    if err != nil {
        panic(err)
    }
    var notes []Note
    for _, directory := range directories {
        if directory.IsDir() == true {
            file, err := os.Open(absPath + "/" + directory.Name() + "/note.html")
            if err != nil {
                panic(err)
            }
            note, err := parseHTML(file)
            if err == nil {
                notes = append(notes, note)
            }
            file.Close()
        }
    }
    return notes
}

func parseHTML(content io.Reader) (Note, error) {
    note := Note{Source: "Catch"}
    tokeniser := html.NewTokenizer(content)
    contentTokenNext := false
    possibleExtIDTokenNext := false
    dateTokenNext := false
    dateScriptTokenNext := false
    for note.Content == "" || note.ExternalID == "" || note.OriginalCreationTimestamp.IsZero() {
        tt := tokeniser.Next()
        switch {
        case tt == html.ErrorToken:
            if note.ExternalID != "" && !note.OriginalCreationTimestamp.IsZero() {
                return note, errors.New("No content")
            }
            panic(
                "Reached end of document without date or externalID" +
                "Timestamp: {{.note.OriginalCreationTimestamp.String()}}" +
                "External Id: {{.note.ExternalID}}" +
                "Content: {{.note.Content}}")
        case tt == html.StartTagToken:
            parseStartTagToken(
                tokeniser,
                &dateTokenNext,
                &possibleExtIDTokenNext,
                &dateScriptTokenNext,
                &contentTokenNext,
            )
        case tt == html.TextToken:
            parseHTMLTextToken(
                tokeniser,
                &note,
                &dateTokenNext,
                &possibleExtIDTokenNext,
                &dateScriptTokenNext,
                &contentTokenNext,
            )
        }
    }
    return note, nil
}

func parseStartTagToken(
    tokeniser *html.Tokenizer,
    dateTokenNext *bool,
    possibleExtIDTokenNext *bool,
    dateScriptTokenNext *bool,
    contentTokenNext *bool,
) {
    token := tokeniser.Token()
    if token.Data == "script" && *dateScriptTokenNext {
        *dateTokenNext = true
    } else {
        for _, a := range token.Attr {
            if a.Key == "class" && a.Val == "note-text" {
                *contentTokenNext = true
            } else if a.Key == "class" && a.Val == "header" {
                *possibleExtIDTokenNext = true
            }
        }
    }
}

func parseDateToken(
    text string,
    dateTokenNext *bool,
    dateScriptTokenNext *bool,
) time.Time {
    timeString := strings.TrimSuffix(strings.TrimPrefix(text, "catch_date("), ")")
    timeInt, err := strconv.ParseInt(timeString, 10, 64)
    if err != nil {
        panic(err)
    }

    *dateTokenNext = false
    *dateScriptTokenNext = false
    return time.Unix(timeInt/1000, 0)
}

func parseHTMLTextToken(
    tokeniser *html.Tokenizer,
    note *Note,
    dateTokenNext *bool,
    possibleExtIDTokenNext *bool,
    dateScriptTokenNext *bool,
    contentTokenNext *bool,
) {
    token := tokeniser.Token()
    text := html.UnescapeString(token.String())
    if *dateTokenNext {
        note.OriginalCreationTimestamp = parseDateToken(text, dateTokenNext, dateScriptTokenNext)
    } else if *possibleExtIDTokenNext && text == "Date: " && note.OriginalCreationTimestamp.IsZero() {
        *dateScriptTokenNext = true
        *possibleExtIDTokenNext = false
    } else if *contentTokenNext {
        note.Content = text
        note.Tags = parseTags(text)
        *contentTokenNext = false
    } else if *possibleExtIDTokenNext {
        if strings.Contains(text, "Note Id: ") {
            note.ExternalID = strings.TrimPrefix(text, "Note Id: ")
        }
        *possibleExtIDTokenNext = false
    }
}

func parseTags(tagContent string) []Tag {
    var tagList []Tag
    for _, unparsedTag := range strings.Split(tagContent, "#")[1:] {
        parsedTag := strings.ToLower(
            strings.Split(
                strings.Split(unparsedTag, "\n")[0],
                " ",
            )[0])
        if parsedTag != "" {
            tagObject := Tag{Name: parsedTag}
            tagList = append(tagList, tagObject)
        }
    }
    return tagList
}

func addExistingTags(
    db *gorm.DB,
    note Note,
    noteIds *[]string,
    tagNames *[]string,
) {
    for _, tag := range note.Tags {
        addExistingTag(db, tag, note, noteIds, tagNames)
    }
}

func addExistingTag(
    db *gorm.DB,
    tag Tag,
    note Note,
    noteIds *[]string,
    tagNames *[]string,
) {
    note.Tags = nil
    if contains(*tagNames, tag.Name) && !contains(*noteIds, note.ExternalID) {
        var tagToMapToNote Tag
        db.Where(&Tag{Name: tag.Name}).Find(&tagToMapToNote)
        err := db.Model(&tagToMapToNote).Association("Notes").Append(&note).Error
        if err != nil {
            panic(err.Error())
        }
        *noteIds = append(*noteIds, note.ExternalID)
    }
}

func ingestData(notes []Note, db *gorm.DB) {
    tx := db.Begin()
    defer tx.Commit()

    var existingTags []Tag
    var tagNames []string
    db.Find(&existingTags).Pluck("name", &tagNames)

    var existingNotes []Note
    var noteIds []string
    db.Find(&existingNotes).Pluck("external_id", &noteIds)

    for _, note := range notes {
        addExistingTags(db, note, &noteIds, &tagNames)
        addNote(db, note, &noteIds, &tagNames)
    }
}

func addNote(
    db *gorm.DB,
    note Note,
    noteIds *[]string,
    tagNames *[]string,
) {
    if !contains(*noteIds, note.ExternalID) {
        err := db.Create(&note).Error
        if err != nil {
            panic(err.Error())
        }
        *noteIds = append(*noteIds, note.ExternalID)
        for _, tag := range note.Tags {
            *tagNames = append(*tagNames, tag.Name)
        }
    }
}

func queryData(db *gorm.DB) []Note {
    var notes []Note
    db.Preload("Tags").Find(&notes)
    return notes
}
