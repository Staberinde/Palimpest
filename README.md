# Welcome to Palimpest!

## Dependencies:
* Postgres 9.6 (earlier versions may work, but are untested)

## Running Palimpest:
Currently Palimpest can only ingest information in the default format exported by Catch Notes
* Install Postgres on your local system, and create a database called `palimpest`
* Optional: Set the environment variable `CATCH_NOTES` to the directory where the exported notes have been unpacked
* Run `go install github.com/Staberinde/Palimpest`
* Run `Palimpest {notes_location}` where {notes_location} is the absolute path to the unpacked exported notes directory. This is only a mandatory argument if the `CATCH_NOTES` environment variable has not been set.

## Running tests:
* Instantiate a postgres database called `test_palimpest` on your local machine
* Run `go test github.com/Staberinde/Palimpest`
