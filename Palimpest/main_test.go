package main

import (
    "os"
    "testing"
    "time"

    "github.com/cloudfoundry/jibber_jabber"
    "github.com/stretchr/testify/assert"
)

func (result *BaseModel) zeroBaseModelDatesAndID() {
    result.ID = 0
    result.CreatedAt = time.Unix(int64(0), int64(0))
    result.UpdatedAt = time.Unix(int64(0), int64(0))
}

func zeroNotesDatesAndID(results []Note) []Note {
    for i, _ := range results {
        results[i].zeroNoteDatesAndID()
    }
    return results
}

func (result *Note) zeroNoteDatesAndID() {
    result.zeroBaseModelDatesAndID()
    for i, _ := range result.Tags {
        result.Tags[i].zeroBaseModelDatesAndID()
    }
}

func getSystemLocalLoc() *time.Location {
    // The representation of a postgres timestamp returned by the pg library is a time.Time with a
    // Location corresponding to the country of the systems current locale, whereas when instantiating
    // a time explicitly in golang by default the Location will be the golangs own abstraction for the
    //  systems current locale. This means two time.Time structs instantiated with simialar parametes
    // will fail comparison if one is inserted and retrieved from a postgres database, so we discover
    // the systems current locale (we make the assumption that this is what postgres is using as its
    // locale) and we fetch the corresponding Location and transform any time.Time structs in expected
    // results to that location
    localeTerritory, err := jibber_jabber.DetectTerritory()
    if err != nil {
        panic(err)
    }
    localLoc, err := time.LoadLocation(localeTerritory)
    if err != nil {
        panic(err)
    }
    return localLoc
}

func TestEndToEnd(t *testing.T) {
    expectedResults := []Note{
        Note{
            Content: `
Welcome to Catch.com
Catch helps you create, organize and sync notes between the web and your mobile devices.
Create it!
Let your mind take a break and offload your thoughts, findings, and images with us.
Organize it!
Simply prefix any word in a note with # to make it a tag: #example
Sync It!
Press the Settings button from the notes list to sign in or create a new Catch account. It's quick, free, and secure.
Your notes, anywhere, anytime: https://catch.com
email: feedback@catch.com
twitter: http://twitter.com/catch
#Welcome #Catch
`,
            Source:                    "Catch",
            ExternalID:                "000000000000000022214014",
            OriginalCreationTimestamp: time.Unix(1295218624, 0).In(getSystemLocalLoc()),
            Tags: []Tag{
                Tag{
                    Name:  "example",
                    Notes: nil,
                },
                Tag{
                    Name:  "welcome",
                    Notes: nil,
                },
                Tag{
                    Name:  "catch",
                    Notes: nil,
                },
            },
        },
    }
    db := setupDatabase(
        "palimpest_test",
        os.Getenv("DATABASE_HOST"),
        os.Getenv("DATABASE_USER"),
        os.Getenv("DATABASE_PASSWORD"),
    )
    db.Exec("truncate notes CASCADE; truncate tags CASCADE; truncate note_tag_mapping CASCADE;")
    defer db.Close()
    openDataAndIngest(db, os.Getenv("TEST_FIXTURES_DIR"))
    results := queryData(db)
    assert.Equal(t, zeroNotesDatesAndID(expectedResults), zeroNotesDatesAndID(results))
}
