package cdrom

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/mitchellh/hashstructure"
	"go.uber.org/multierr"
)

// Column describes a CDROM instance from the procfs cdrom/info table file
type Column struct {
	ID string

	Speed uint64
	Slots uint64

	CanChangeSpeed  bool
	CanSelectDisk   bool
	CanMultiSession bool
	CanMediaChanged bool
	CanReadMCN      bool
	CanWriteCDR     bool
	CanWriteCDRW    bool
	CanWriteRAM     bool
	CanReadDVD      bool
	CanWriteDVDR    bool
	CanWriteDVDRAM  bool
	CanReadMRW      bool
	CanWriteMRW     bool
}

// Table row headers
const (
	DriveNameHeader = "drive name:"

	DriveSpeed           = "drive speed"
	DriveSlots           = "drive # of slots"
	DriveCanChangeSpeed  = "Can change speed"
	DriveCanSelectDisk   = "Can select disk"
	DriveCanMediaChanged = "Reports media changed"
	DriveCanMultiSession = "Can read multisession"
	DriveCanReadMCN      = "Can read MCN"
	DriveCanWriteCDR     = "Can write CD-R"
	DriveCanWriteCDRW    = "Can write CD-RW"
	DriveCanReadDVD      = "Can read DVD"
	DriveCanWriteDVDR    = "Can write DVD-R"
	DriveCanWriteDVDRAM  = "Can write DVD-RAM"
	DriveCanReadMRW      = "Can read MRW"
	DriveCanWriteMRW     = "Can write MRW"
	DriveCanWriteRAM     = "Can write RAM"
)

// LoadTable reads a procfs cdrom/info table into a map of Device structs
func LoadTable(file io.Reader) (columns []Column, fingerprint string, errs error) {
	scanner := bufio.NewScanner(file)

	// Find the `drive name:` table row and initialize columns
	columns, errs = ScanHeader(scanner)
	if errs != nil {
		return
	}

	// Scan subsequent lines and map properties into Device entries
	for scanner.Scan() {
		var err error

		property, values, has := strings.Cut(scanner.Text(), ":")
		if !has {
			continue
		}

		fields := strings.Fields(values)
		if len(fields) != len(columns) {
			errs = multierr.Append(errs, fmt.Errorf("invalid column count for property %q: expected %d, parsed %d `%v`", property, len(columns), len(fields), fields))
			continue
		}

		// Parse and set the property for each drive in the devices map
		for i, field := range fields {
			switch property {
			case DriveSpeed:
				columns[i].Speed, err = strconv.ParseUint(field, 10, 64)
			case DriveSlots:
				columns[i].Slots, err = strconv.ParseUint(field, 10, 64)

			case DriveCanChangeSpeed:
				columns[i].CanChangeSpeed, err = strconv.ParseBool(field)
			case DriveCanSelectDisk:
				columns[i].CanSelectDisk, err = strconv.ParseBool(field)
			case DriveCanMediaChanged:
				columns[i].CanMediaChanged, err = strconv.ParseBool(field)
			case DriveCanMultiSession:
				columns[i].CanMultiSession, err = strconv.ParseBool(field)
			case DriveCanReadMCN:
				columns[i].CanReadMCN, err = strconv.ParseBool(field)
			case DriveCanWriteCDR:
				columns[i].CanWriteCDR, err = strconv.ParseBool(field)
			case DriveCanWriteCDRW:
				columns[i].CanWriteCDRW, err = strconv.ParseBool(field)
			case DriveCanReadDVD:
				columns[i].CanReadDVD, err = strconv.ParseBool(field)
			case DriveCanWriteDVDR:
				columns[i].CanWriteDVDR, err = strconv.ParseBool(field)
			case DriveCanWriteDVDRAM:
				columns[i].CanWriteDVDRAM, err = strconv.ParseBool(field)
			case DriveCanReadMRW:
				columns[i].CanReadMRW, err = strconv.ParseBool(field)
			case DriveCanWriteMRW:
				columns[i].CanWriteMRW, err = strconv.ParseBool(field)
			case DriveCanWriteRAM:
				columns[i].CanWriteRAM, err = strconv.ParseBool(field)
			}

			if err != nil {
				errs = multierr.Append(errs, fmt.Errorf("unable to parse value `%s: %s` for column %d", property, field, i))
			}
		}
	}

	if errs != nil {
		return
	}

	// Get a fingerprint for the parsed table for change detection
	fp, errs := hashstructure.Hash(columns, nil)
	if errs != nil {
		return
	}

	fingerprint = strconv.FormatUint(fp, 16)
	return
}

// ScanHeader reads the info file until a line beginning with `drive name:` is reached, then reads drive identifier
// fields and allocates a Device for each
func ScanHeader(scanner *bufio.Scanner) (columns []Column, err error) {
	// Find the first line of the table get device IDs and allocate map entries
	for scanner.Scan() {
		line, match := strings.CutPrefix(scanner.Text(), DriveNameHeader)
		if match {
			// Split `\tsr0\tsr1\t...` into ["sr0", "sr1", ...]
			fields := strings.Fields(line)

			// Initialize entries for each column
			columns = make([]Column, len(fields))
			for i, field := range fields {
				columns[i].ID = field
			}

			break
		}
	}

	err = scanner.Err()
	return
}
