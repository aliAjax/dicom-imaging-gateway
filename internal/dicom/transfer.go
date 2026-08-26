package dicom

import "fmt"

type TransferSyntax string

const (
	ImplicitLittle TransferSyntax = "1.2.840.10008.1.2"
	ExplicitLittle TransferSyntax = "1.2.840.10008.1.2.1"
)

func (t TransferSyntax) Validate() error {
	switch t {
	case ImplicitLittle, ExplicitLittle:
		return nil
	default:
		return fmt.Errorf("transfer_syntax_unsupported")
	}
}
func Negotiated(local, remote []TransferSyntax) (TransferSyntax, error) {
	for _, l := range local {
		for _, r := range remote {
			if l == r {
				return l, nil
			}
		}
	}
	return "", fmt.Errorf("transfer_syntax_no_common")
}
