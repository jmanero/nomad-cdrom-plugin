package device_test

import (
	"os"
	"testing"

	"github.com/jmanero/nomad-cdrom-plugin/device"
	"github.com/stretchr/testify/assert"
)

func TestParseInfo_MultipleColumns(t *testing.T) {
	info, err := os.Open("testdata/info0.txt")
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	defer info.Close()

	devices, fingerprint, err := device.LoadTable(info)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	assert.Equal(t, "83f6e601f6255ce7", fingerprint)
	assert.ElementsMatch(t, []device.Column{
		{
			ID:              "sr0",
			Speed:           0,
			Slots:           1,
			CanChangeSpeed:  true,
			CanMultiSession: true,
			CanMediaChanged: true,
			CanReadMCN:      true,
			CanReadDVD:      true,
		},
		{
			ID:              "sr1",
			Speed:           16,
			Slots:           1,
			CanChangeSpeed:  true,
			CanMultiSession: true,
			CanMediaChanged: true,
			CanReadMCN:      true,
			CanWriteCDR:     true,
			CanWriteCDRW:    true,
		},
	}, devices)
}

func TestParseInfo_Properties(t *testing.T) {
	info, err := os.Open("testdata/info1.txt")
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	defer info.Close()

	devices, fingerprint, err := device.LoadTable(info)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	assert.Equal(t, "7a179ef5c41fe8ea", fingerprint)
	assert.ElementsMatch(t, devices, []device.Column{
		{
			ID:              "sr0",
			Speed:           48,
			Slots:           4,
			CanChangeSpeed:  true,
			CanSelectDisk:   true,
			CanMultiSession: true,
			CanMediaChanged: true,
			CanReadMCN:      true,
			CanWriteCDR:     true,
			CanWriteCDRW:    true,
			CanReadDVD:      true,
			CanWriteDVDR:    true,
			CanWriteDVDRAM:  true,
			CanReadMRW:      true,
			CanWriteMRW:     true,
			CanWriteRAM:     true,
		},
	})
}
