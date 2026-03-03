package document

// Table is a document node that arranges content in a grid of rows and
// columns. It supports separate header, body, and footer sections, as
// well as column width definitions and cell spanning.
type Table struct {
	// Columns defines the width of each column.
	Columns []TableColumn
	// Header contains rows that repeat at the top of each page when the
	// table spans multiple pages.
	Header []TableRow
	// Body contains the main content rows.
	Body []TableRow
	// Footer contains rows that repeat at the bottom of each page when
	// the table spans multiple pages.
	Footer []TableRow
	// TableStyle controls the table's visual properties.
	TableStyle TableStyle
}

// TableColumn defines properties for a single table column.
type TableColumn struct {
	// Width specifies the column width. Use Auto for automatic sizing.
	Width Value
}

// TableRow represents a horizontal row of cells within a table section.
type TableRow struct {
	// Cells holds the cells in this row.
	Cells []TableCell
}

// TableCell represents a single cell in a table. It can span multiple
// columns and rows.
type TableCell struct {
	// Content holds the child nodes rendered inside this cell.
	Content []DocumentNode
	// ColSpan is the number of columns this cell spans (minimum 1).
	ColSpan int
	// RowSpan is the number of rows this cell spans (minimum 1).
	RowSpan int
	// CellStyle controls the cell's visual properties.
	CellStyle Style
}

// TableStyle extends BoxStyle with table-specific visual properties.
type TableStyle struct {
	BoxStyle
	// BorderCollapse when true merges adjacent cell borders into a single
	// border, similar to CSS border-collapse: collapse.
	BorderCollapse bool
}

// NodeType returns NodeTable.
func (t *Table) NodeType() NodeType { return NodeTable }

// Children collects all cells across header, body, and footer rows into
// a flat list of DocumentNode values. This is primarily used for tree
// traversal; the table layout engine uses the structured row/cell data
// directly.
func (t *Table) Children() []DocumentNode {
	var children []DocumentNode
	for _, row := range t.Header {
		for i := range row.Cells {
			children = append(children, &CellNode{Cell: &row.Cells[i]})
		}
	}
	for _, row := range t.Body {
		for i := range row.Cells {
			children = append(children, &CellNode{Cell: &row.Cells[i]})
		}
	}
	for _, row := range t.Footer {
		for i := range row.Cells {
			children = append(children, &CellNode{Cell: &row.Cells[i]})
		}
	}
	return children
}

// Style returns a Style derived from the table's BoxStyle.
func (t *Table) Style() Style {
	return Style{
		Margin:     t.TableStyle.Margin,
		Padding:    t.TableStyle.Padding,
		Border:     t.TableStyle.Border,
		Background: t.TableStyle.Background,
	}
}

// CellNode wraps a TableCell to implement the DocumentNode interface,
// allowing cells to participate in generic tree traversal.
type CellNode struct {
	Cell *TableCell
}

// NodeType returns NodeTableCell.
func (cn *CellNode) NodeType() NodeType { return NodeTableCell }

// Children returns the content nodes inside this cell.
func (cn *CellNode) Children() []DocumentNode { return cn.Cell.Content }

// Style returns the cell's visual style.
func (cn *CellNode) Style() Style { return cn.Cell.CellStyle }
