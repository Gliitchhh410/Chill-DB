package domain


type ColumnDefinition struct {
	Name string
	Type string
}


type TableMetaData struct {
	Name string
	Columns []ColumnDefinition
}

type Row []string