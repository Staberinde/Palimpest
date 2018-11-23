package main

import (
    "time"
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/cloudfoundry/jibber_jabber"
)

func (result *BaseModel) zeroBaseModelDatesAndID() {
    result.ID = 0
    result.CreatedAt = time.Time{}
    result.UpdatedAt = time.Time{}
}

func zeroNotesDatesAndID(results []Note) {
    for _, result := range results {
        result.zeroNoteDatesAndID()
    }
}

func (result *Note) zeroNoteDatesAndID(){
    result.zeroBaseModelDatesAndID()
    for _, Tag := range result.Tags {
        Tag.zeroBaseModelDatesAndID()
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
            Source: "Catch",
            ExternalID: "000000000000000022214014",
            OriginalCreationTimestamp: time.Unix(1295218624, 0).In(getSystemLocalLoc()),
            Tags: []Tag{
                Tag{
                    Name: "example",
                    Notes: nil,
                },
                Tag{
                    Name: "welcome",
                    Notes: nil,
                },
                Tag{
                    Name: "catch",
                    Notes: nil,
                },
            },
        },
    }
    zeroNotesDatesAndID(expectedResults)
    db := setupDatabase("palimpest_test")
    db.Exec("truncate notes CASCADE; truncate tags CASCADE; truncate note_tag_mapping CASCADE;")
    defer db.Close()
    openDataAndIngest(db, "./TestFixtures")
    results := queryData(db)
    zeroNotesDatesAndID(results)
    assert.Equal(t, expectedResults, results)
}
