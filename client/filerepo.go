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
    "bytes"
    "os"
_   "fmt"
_   "log"
    "time"
    "path/filepath"
    "strings"
    "regexp"
)

// The default metadata format.
var DefaultFormat Format = Format{
    Prefix:     "iso19139",
    Schema:     "http://www.isotc211.org/2005/gmi/gmi.xsd",
    Namespace:  "http://www.isotc211.org/2005/gmi",
}

// Pattern to use selecting XML processing instructions
var xmlPIRegExp *regexp.Regexp = regexp.MustCompile(`<\?[^?]*\?>`)


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
    // If no set is specific, scan all the sets
    if (set == "") {
        sets, err := fr.Sets()
        if (err != nil) {
            return nil, err
        }

        allRecords := make([]*Record, 0)
        for _, aset := range sets {
            recs, err := fr.scanRecordsFromDir(aset.Spec, func(rec *Record) bool { return true })
            if (err != nil) {
                return nil, err
            }
            allRecords = append(allRecords, recs...)
        }

        return &SliceRecordCursor{allRecords, 0}, nil

    } else {
        recs, err := fr.scanRecordsFromDir(set, func(rec *Record) bool { return true })
        if (err != nil) {
            return nil, err
        }

        return &SliceRecordCursor{recs, 0}, nil
    }
}

// Returns a record
func (fr *FileRepository) Record(id string) (*Record, error) {
    // Search for the record in each of the sets in turn
    sets, err := fr.Sets()
    if (err != nil) {
        return nil, err
    }

    for _, set := range sets {
        setname := set.Spec
        record := fr.readRecordFromSet(setname, id)
        if (record != nil) {
            return record, nil
        }
    }
    return nil, nil
}

// Scan for "metadata" records from a directory.
func (fr *FileRepository) scanRecordsFromDir(setname string, filter func(rec *Record) bool) ([]*Record, error) {
    dirName := filepath.Join(fr.BaseDir, setname)
    dir, err := os.Open(dirName)
    if (err != nil) {
        return nil, err
    }
    defer dir.Close()

    // Read the contents of the directory
    files, err := dir.Readdir(-1)
    if (err != nil) {
        return nil, err
    }

    // Convert them into records
    recs := make([]*Record, 0, len(files))
    for _, file := range files {
        fullFilename := filepath.Join(dirName, file.Name())
        rec := fr.buildRecord(setname, fullFilename, file)
        if (rec != nil) && (filter(rec)) {
            recs = append(recs, rec)
        }
    }

    return recs, nil
}

// Attempts to load a record from a set
func (fr *FileRepository) readRecordFromSet(set string, id string) *Record {
    basename := id + ".xml"
    recordPath := filepath.Join(fr.BaseDir, set, basename)
    fileInfo, err := os.Stat(recordPath)
    if (err == nil) {
        return fr.buildRecord(set, recordPath, fileInfo)
    } else {
        return nil
    }
}

// Build a record from a file info.  Returns nil if a record cannot be built from a file.
func (fr *FileRepository) buildRecord(set string, filename string, fileInfo os.FileInfo) *Record {
    basename := fileInfo.Name()

    if (! strings.HasSuffix(basename, ".xml")) {
        return nil
    }
    trimmedFilename := strings.TrimSuffix(basename, ".xml")

    return &Record{
        ID: trimmedFilename,
        Date: fileInfo.ModTime(),
        Set: set,
        Content: func() (string, error) {
            file, err := os.Open(filename)

            if (err != nil) {
                return "", err
            }
            defer file.Close()

            buffer := new(bytes.Buffer)
            buffer.ReadFrom(file)
            content := buffer.String()

            // Remove processing instructions
            content = xmlPIRegExp.ReplaceAllString(content, "")
            return content, nil
        },
    }
}

// --------------------------------------------------------------------------------
// A cursor for navigating a slice

type SliceRecordCursor struct {
    Records     []*Record
    Pointer     int
}

// Returns true if the particular position is valid
func (c *SliceRecordCursor) posValid(p int) bool {
    return (p >= 0) && (p < len(c.Records))
}

// Indicates if the cursor has more records
func (c *SliceRecordCursor) HasRecord() bool {
    return c.posValid(c.Pointer)
}

// Goes to the next record.  If the next record exists, returns true.  Otherwise, returns false.
func (c *SliceRecordCursor) Next() bool {
    c.Pointer++
    if (c.posValid(c.Pointer)) {
        return true
    } else {
        return false
    }
}

// Moves the cursor to a particular position.  If the position is valid, returns true.
func (c *SliceRecordCursor) SetPos(pos int) bool {
    if (c.posValid(pos)) {
        c.Pointer = pos
        return true
    } else {
        return false
    }
}

// Returns the current position of the cursor.
func (c *SliceRecordCursor) Pos() int {
    return c.Pointer
}

// Returns the current record, or nil if the cursor is at an invalid position.
func (c *SliceRecordCursor) Record() *Record {
    if (c.posValid(c.Pointer)) {
        return c.Records[c.Pointer]
    } else {
        return nil
    }
}
