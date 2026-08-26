package dicom

import "time"

type Tag struct {
	Group   uint16 `json:"group"`
	Element uint16 `json:"element"`
}

func (t Tag) String() string { return fmtTag(t.Group, t.Element) }
func fmtTag(g, e uint16) string {
	const hex = "0123456789ABCDEF"
	b := make([]byte, 9)
	b[0] = '('
	b[5] = ','
	b[8] = ')'
	for i := 0; i < 4; i++ {
		b[1+i] = hex[(g>>uint(12-4*i))&15]
		b[6+i] = hex[(e>>uint(12-4*i))&15]
	}
	return string(b)
}

type Element struct {
	Tag       Tag      `json:"tag"`
	VR        string   `json:"vr"`
	Length    uint32   `json:"length"`
	Value     []byte   `json:"-"`
	Fragments [][]byte `json:"-"`
}

func (e Element) Text() string { return string(e.Value) }

type Dataset struct {
	TransferSyntax string    `json:"transferSyntax"`
	Elements       []Element `json:"elements"`
	StudyUID       string    `json:"studyUID"`
	SeriesUID      string    `json:"seriesUID"`
	SOPInstanceUID string    `json:"sopInstanceUID"`
	PatientName    string    `json:"patientName"`
	PatientID      string    `json:"patientID"`
	Modality       string    `json:"modality"`
	ReceivedAt     time.Time `json:"receivedAt"`
	SHA256         string    `json:"sha256"`
	Size           int       `json:"size"`
}
type Instance struct {
	UID       string    `json:"uid"`
	StudyUID  string    `json:"studyUID"`
	SeriesUID string    `json:"seriesUID"`
	Metadata  Dataset   `json:"metadata"`
	ObjectKey string    `json:"objectKey"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"createdAt"`
	Version   int64     `json:"version"`
}
type Destination struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	AETitle       string `json:"aeTitle"`
	Endpoint      string `json:"endpoint"`
	Enabled       bool   `json:"enabled"`
	MaxConcurrent int    `json:"maxConcurrent"`
}
type RouteJob struct {
	ID            string    `json:"id"`
	InstanceUID   string    `json:"instanceUID"`
	DestinationID string    `json:"destinationID"`
	Status        string    `json:"status"`
	Attempts      int       `json:"attempts"`
	LastError     string    `json:"lastError,omitempty"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}
