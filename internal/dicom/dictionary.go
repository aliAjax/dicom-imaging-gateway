package dicom

import "strings"

type DictionaryEntry struct {
	Tag     Tag    `json:"tag"`
	Keyword string `json:"keyword"`
	Name    string `json:"name"`
	VR      string `json:"vr"`
	Retired bool   `json:"retired"`
}

var standardEntries = []DictionaryEntry{
	{Tag{0x0002, 0x0000}, "FileMetaInformationGroupLength", "File Meta Information Group Length", "UL", false},
	{Tag{0x0002, 0x0001}, "FileMetaInformationVersion", "File Meta Information Version", "OB", false},
	{Tag{0x0002, 0x0002}, "MediaStorageSOPClassUID", "Media Storage SOP Class UID", "UI", false},
	{Tag{0x0002, 0x0003}, "MediaStorageSOPInstanceUID", "Media Storage SOP Instance UID", "UI", false},
	{Tag{0x0002, 0x0010}, "TransferSyntaxUID", "Transfer Syntax UID", "UI", false},
	{Tag{0x0002, 0x0012}, "ImplementationClassUID", "Implementation Class UID", "UI", false},
	{Tag{0x0008, 0x0016}, "SOPClassUID", "SOP Class UID", "UI", false},
	{Tag{0x0008, 0x0018}, "SOPInstanceUID", "SOP Instance UID", "UI", false},
	{Tag{0x0008, 0x0020}, "StudyDate", "Study Date", "DA", false},
	{Tag{0x0008, 0x0030}, "StudyTime", "Study Time", "TM", false},
	{Tag{0x0008, 0x0060}, "Modality", "Modality", "CS", false},
	{Tag{0x0008, 0x0090}, "ReferringPhysicianName", "Referring Physician Name", "PN", false},
	{Tag{0x0010, 0x0010}, "PatientName", "Patient Name", "PN", false},
	{Tag{0x0010, 0x0020}, "PatientID", "Patient ID", "LO", false},
	{Tag{0x0010, 0x0030}, "PatientBirthDate", "Patient Birth Date", "DA", false},
	{Tag{0x0010, 0x0040}, "PatientSex", "Patient Sex", "CS", false},
	{Tag{0x0020, 0x000d}, "StudyInstanceUID", "Study Instance UID", "UI", false},
	{Tag{0x0020, 0x000e}, "SeriesInstanceUID", "Series Instance UID", "UI", false},
	{Tag{0x0020, 0x0011}, "SeriesNumber", "Series Number", "IS", false},
	{Tag{0x0020, 0x0013}, "InstanceNumber", "Instance Number", "IS", false},
	{Tag{0x0028, 0x0010}, "Rows", "Rows", "US", false},
	{Tag{0x0028, 0x0011}, "Columns", "Columns", "US", false},
	{Tag{0x0028, 0x0100}, "BitsAllocated", "Bits Allocated", "US", false},
	{Tag{0x0028, 0x0101}, "BitsStored", "Bits Stored", "US", false},
	{Tag{0x0028, 0x0103}, "PixelRepresentation", "Pixel Representation", "US", false},
	{Tag{0x7fe0, 0x0010}, "PixelData", "Pixel Data", "OW", false},
}

type Dictionary struct {
	byTag     map[Tag]DictionaryEntry
	byKeyword map[string]DictionaryEntry
}

func NewDictionary() *Dictionary {
	d := &Dictionary{byTag: map[Tag]DictionaryEntry{}, byKeyword: map[string]DictionaryEntry{}}
	for _, e := range standardEntries {
		d.byTag[e.Tag] = e
		d.byKeyword[strings.ToLower(e.Keyword)] = e
	}
	return d
}
func (d *Dictionary) Lookup(tag Tag) (DictionaryEntry, bool) { e, ok := d.byTag[tag]; return e, ok }
func (d *Dictionary) Find(keyword string) (DictionaryEntry, bool) {
	e, ok := d.byKeyword[strings.ToLower(strings.TrimSpace(keyword))]
	return e, ok
}
func IsPrivate(tag Tag) bool     { return tag.Group%2 == 1 }
func IsGroupLength(tag Tag) bool { return tag.Element == 0x0000 }
func IsUIDVR(vr string) bool     { return vr == "UI" }
