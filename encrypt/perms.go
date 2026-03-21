package encrypt

// Permission represents PDF document permission flags (ISO 32000-2 Table 22).
type Permission uint32

const (
	// PermPrint allows printing the document.
	PermPrint Permission = 1 << 2
	// PermModify allows modifying the document contents.
	PermModify Permission = 1 << 3
	// PermCopy allows copying text and graphics from the document.
	PermCopy Permission = 1 << 4
	// PermAnnotate allows adding or modifying annotations.
	PermAnnotate Permission = 1 << 5
	// PermFillForms allows filling in form fields.
	PermFillForms Permission = 1 << 8
	// PermExtract allows extracting text and graphics for accessibility.
	PermExtract Permission = 1 << 9
	// PermAssemble allows assembling the document (insert, rotate, delete pages).
	PermAssemble Permission = 1 << 10
	// PermPrintHighRes allows printing at high resolution.
	PermPrintHighRes Permission = 1 << 11

	// PermAll grants all permissions.
	PermAll = PermPrint | PermModify | PermCopy | PermAnnotate |
		PermFillForms | PermExtract | PermAssemble | PermPrintHighRes
)
