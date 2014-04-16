// Implementation of the repository based on the file system.
// The format that this file-system based repository must adhere to is:
//
//      basedir
//          setname
//              metadataId.xml
//
// The metadata ID will be everything before the extension.  The modification
// date of the file will be used as the metadata date.
//

package oaipmh

import (
    "os"
    "fmt"
    "time"
)

// The default metadata format.
var DefaultFormat Format = Format{
    Prefix:     "iso19139",
    Schema:     "http://www.isotc211.org/2005/gmi/gmi.xsd",
    Namespace:  "http://www.isotc211.org/2005/gmi",
}


// A file based repository
type FileRepository struct {

    // Path to the base directory
    BaseDir         string

    // The format that this repository manages.
    Format          Format
}

// Creates a new FileRepository with the format set to the default format.
func NewFileRepository(basedir string) *FileRepository {
    return &FileRepository{basedir, DefaultFormat}
}


// Simply returns the passed in Format.
func (fr *FileRepository) Formats() []Format {
    return []Format { fr.Format }
}

// Returns the sets managed by the repository.  These will be the directories that exist
// directly underneith the base directory.
func (fr *FileRepository) Sets() ([]Set, error) {
    dir, err := os.Open(fr.BaseDir)
    if (err != nil) {
        return nil, err
    }
    defer dir.Close()

    // Read all the directories
    subdirs, err := dir.Readdir(-1)
    if (err != nil) {
        return nil, err
    }

    // Convert them into sets
    sets := make([]Set, 0, len(subdirs))
    for _, subdir := range subdirs {
        if (subdir.IsDir()) {
            sets = append(sets, Set{
                Spec: subdir.Name(),
                Name: subdir.Name(),
                Descr: "",
            })
        }
    }

    return sets, nil
}


// Reads the record from a set.  This simply iterates over all the files in a set directory.
func (fr *FileRepository) ListRecords(set string, from time.Time, to time.Time) (RecordCursor, error) {
    return nil, fmt.Errorf("Not implemented yet")
}

// Returns a record
func (fr *FileRepository) Record(id string) (*Record, error) {
    return nil, fmt.Errorf("Not implemented yet")
}
