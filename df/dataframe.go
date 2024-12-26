package df

import (
	"fmt"
	"math/rand"
	"slices"
	"strings"

	"github.com/discoverkl/goterm/df/vs"
)

// SupportedType constrains the types that can be used in a Series
type SupportedType interface {
	~string | ~float64 | ~int
}

type Series interface {
	fmt.Stringer

	Len() int
	Name() string
	Data() []any
	ToFloat64() []float64
	AsFloat64() []float64
	AsInt() []int
	AsString() []string
	Avg() Series
}

// Concrete implementation for Series
type series struct {
	name string
	data []any
}

func (s *series) Len() int {
	return len(s.data)
}

func (s *series) Name() string {
	return s.name
}

func (s *series) Data() []any {
	return s.data
}

func (s *series) ToFloat64() []float64 {
	size := len(s.data)
	if size == 0 {
		return []float64{}
	}
	switch s.data[0].(type) {
	case float64:
		return Map(s.data, func(v any) float64 {
			return float64(v.(float64))
		})
	case int:
		return Map(s.data, func(v any) float64 {
			return float64(v.(int))
		})
	case string:
		return slices.Collect(vs.IntRange(0, size-1))
	default:
		return make([]float64, size)
	}
}

func (s *series) AsFloat64() []float64 {
	return Map(s.data, func(v any) float64 {
		return v.(float64)
	})
}

func (s *series) AsInt() []int {
	return Map(s.data, func(v any) int {
		return v.(int)
	})
}

func (s *series) AsString() []string {
	return Map(s.data, func(v any) string {
		return v.(string)
	})
}

func (s *series) Avg() Series {
	if len(s.data) == 0 {
		return NewSeries("avg", []float64{})
	}
	var avg float64
	switch s.data[0].(type) {
	case float64:
		avg = Avg(s.AsFloat64())
	case int:
		avg = Avg(s.AsInt())
	case string:
		return NewSeries(s.name, []string{"Avg"})
	default:
		panic("unsupported")
	}
	return NewSeries(s.name, []float64{avg})
}

func (s *series) String() string {
	index := []int{}
	for i := 0; i < s.Len(); i++ {
		index = append(index, i)
	}
	indexCol := NewSeries("index", index)
	df := NewDataFrame(indexCol, s)
	return df.String()
}

func NewSeries[T SupportedType](name string, data []T) Series {
	// get the type of the first element
	return &series{
		name: name,
		data: AsAny(data),
	}
}

func NewSeriesAny(name string, data []any) Series {
	if len(data) > 0 {
		switch data[0].(type) {
		case float64, int, string:
		default:
			panic("unsupported")
		}
	}
	return &series{
		name: name,
		data: data,
	}
}

func NewRandomIntSeries(name string, len, max int) Series {
	if len < 0 {
		panic("len cannot be negative")
	}
	if max <= 0 {
		max = 100
	}

	data := make([]int, len)
	for i := range data {
		data[i] = rand.Intn(max)
	}
	return NewSeries(name, data)
}

func NewRandomFloat64Series(name string, len int, min float64, max float64) Series {
	if len < 0 {
		panic("len cannot be negative")
	}
	if min >= max {
		panic("min must be less than max")
	}

	data := make([]float64, len)
	for i := range data {
		data[i] = rand.Float64()*(max-min) + min
	}
	return NewSeries(name, data)
}

func NewStringSeries(name string, len int) Series {
	if len < 0 {
		panic("len cannot be negative")
	}

	data := make([]string, len)
	for i := range data {
		data[i] = convertToBase26String(i + 1)
	}
	return NewSeries(name, data)
}

func convertToBase26String(n int) string {
	if n <= 0 {
		panic("n cannot be negative")
	}

	var res strings.Builder
	for n > 0 {
		n--
		res.WriteByte(byte('A' + n%26))
		n /= 26
	}
	// reverse the string
	resStr := res.String()
	res.Reset()
	for i := len(resStr) - 1; i >= 0; i-- {
		res.WriteByte(resStr[i])
	}
	return res.String()
}

// DataFrame interface to define the operations on a DataFrame
type DataFrame interface {
	fmt.Stringer

	Columns() []string
	Rows() int

	GetColumn(name string) Series
	GetColumnAt(index int) Series
	SetColumn(data Series) error
	SetColumnAt(index int, data Series) error
	RemoveColumn(name string) error
	RemoveColumnAt(index int) error

	Head(n int) DataFrame
	Tail(n int) DataFrame
	Avg() DataFrame

	// Plot(options ...ChartOption)
	Bar(options ...ChartOption)
	Line(options ...ChartOption)
	Pie(options ...ChartOption)
	XY(options ...ChartOption)
}

// Concrete implementation for DataFrame
type dataFrame struct {
	columns map[string]Series // Data for columns
	order   []string          // Order of columns
}

// NewDataFrame creates a new DataFrame with the given columns in the given order
func NewDataFrame(columns ...Series) DataFrame {
	df := &dataFrame{
		columns: make(map[string]Series),
		order:   make([]string, 0, len(columns)),
	}

	for _, col := range columns {
		df.SetColumn(col)
	}
	return df
}

// Helper function to check if a string slice contains a string
func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}

func (df *dataFrame) Columns() []string {
	// Return the column names in the order they were added
	return append([]string{}, df.order...)
}

func (df *dataFrame) Rows() int {
	if len(df.columns) == 0 {
		return 0
	}
	for _, colName := range df.order {
		if s, ok := df.columns[colName]; ok {
			return s.Len()
		}
	}
	return 0 // This should not happen if implemented correctly
}

func (df *dataFrame) GetColumn(name string) Series {
	index := slices.Index(df.order, name)
	if index == -1 {
		return nil
	}
	return df.GetColumnAt(index)
}

func (df *dataFrame) GetColumnAt(index int) Series {
	if index < 0 || index >= len(df.order) {
		return nil
	}
	name := df.order[index]
	if col, exists := df.columns[name]; exists {
		return col
	}
	return nil
}

func (df *dataFrame) SetColumn(data Series) error {
	name := data.Name()
	df.columns[name] = data

	// If the column name is not in order, add it to the end
	if !contains(df.order, name) {
		df.order = append(df.order, name)
	}
	return nil
}

func (df *dataFrame) SetColumnAt(index int, data Series) error {
	if index < 0 || index > len(df.order) {
		return fmt.Errorf("index out of range")
	}
	name := data.Name()
	df.columns[name] = data

	df.order = slices.Insert(df.order, index, name)
	return nil
}

func (df *dataFrame) RemoveColumn(name string) error {
	index := slices.Index(df.order, name)
	if index == -1 {
		return fmt.Errorf("column not found")
	}
	return df.RemoveColumnAt(index)
}

func (df *dataFrame) RemoveColumnAt(index int) error {
	if index < 0 || index >= len(df.order) {
		return fmt.Errorf("index out of range")
	}
	name := df.order[index]
	delete(df.columns, name)
	df.order = slices.Delete((df.order), index, index+1)
	return nil
}

func (df *dataFrame) Head(n int) DataFrame {
	if n >= df.Rows() {
		return df
	}

	// Create a new DataFrame with the first n rows of each column
	columns := []Series{}
	for _, colName := range df.order {
		s := df.GetColumn(colName)
		columns = append(columns, NewSeriesAny(colName, s.Data()[:n]))
	}
	return NewDataFrame(columns...)
}

func (df *dataFrame) Tail(n int) DataFrame {
	if n >= df.Rows() {
		return df
	}

	// Create a new DataFrame with the last n rows of each column
	columns := []Series{}
	for _, colName := range df.order {
		s := df.GetColumn(colName)
		columns = append(columns, NewSeriesAny(colName, s.Data()[df.Rows()-n:]))
	}
	return NewDataFrame(columns...)
}

func (df *dataFrame) Avg() DataFrame {
	columns := []Series{}
	for i := range df.order {
		s := df.GetColumnAt(i)
		columns = append(columns, s.Avg())
	}
	return NewDataFrame(columns...)
}

// String returns a string representation of the DataFrame
func (df *dataFrame) String() string {
	data := [][]string{}

	// Add the column names as the first row
	data = append(data, df.Columns())

	if df.Rows() == 0 {
		return "<empty DataFrame>"
	}

	// get first row
	row := []any{}
	for _, col := range df.Columns() {
		s := df.GetColumn(col)
		row = append(row, s.Data()[0])
	}

	// get column format strings based on the type of the first row
	colFormats := make([]string, len(row))
	for i, cell := range row {
		switch cell.(type) {
		case float64:
			colFormats[i] = "%.6f"
		case int:
			colFormats[i] = "%d"
		default:
			colFormats[i] = "%s"
		}
	}

	// Add the data rows
	for i := 0; i < df.Rows(); i++ {
		row := []string{}
		for j, col := range df.Columns() {
			s := df.GetColumn(col)
			row = append(row, fmt.Sprintf(colFormats[j], s.Data()[i]))
		}
		data = append(data, row)
	}

	// get max length of each column, including the column name
	colLengths := make([]int, len(data[0]))
	for i, col := range data[0] {
		colLengths[i] = len(col)
	}
	for _, row := range data {
		for i, cell := range row {
			if len(cell) > colLengths[i] {
				colLengths[i] = len(cell)
			}
		}
	}

	// get the format string for every row
	format := ""
	for _, l := range colLengths {
		format += fmt.Sprintf("%%%ds ", l)
	}
	format += "\n"

	// format the data
	var buf strings.Builder
	for _, row := range data {
		var args []any
		for _, cell := range row {
			args = append(args, cell)
		}
		buf.WriteString(fmt.Sprintf(format, args...))
	}
	return strings.TrimRight(buf.String(), "\n")
}

// FromRecords creates a DataFrame from a slice of slices where each inner slice represents a row
func FromRecords(data [][]any, columns []string) DataFrame {
	// if row count is zero, return an empty DataFrame with the given columns
	if len(data) == 0 {
		return &dataFrame{columns: make(map[string]Series), order: columns}
	}

	// Check if the number of columns in the data matches the number of columns provided
	numColumns := len(columns)
	if numColumns != len(data[0]) {
		panic("Number of columns provided doesn't match")
	}

	// Create a DataFrame and populate it with the data
	df := &dataFrame{columns: make(map[string]Series), order: columns}

	// Transpose the data to create columns
	for i := 0; i < numColumns; i++ {
		var colData []any

		// Extract the i-th element from each row
		for _, row := range data {
			if i < len(row) {
				colData = append(colData, row[i])
			}
		}

		// Create a new Series and add it to the DataFrame
		series := NewSeriesAny(columns[i], colData)
		df.SetColumn(series)
	}
	return df
}

// FromRandomValue generates a DataFrame with random float64 values.
func FromRandomValue(rows, cols int, columns []string) DataFrame {
	if len(columns) != cols {
		panic("Number of columns provided doesn't match requested cols")
	}

	data := make([][]any, rows)
	for i := range data {
		data[i] = make([]any, cols)
		for j := range data[i] {
			data[i][j] = randomValue(0.0)
		}
	}
	return FromRecords(data, columns)
}

func randomValue(v any) any {
	switch vtype := (v).(type) {
	case float64:
		return rand.Float64()*2 - 1
	case int:
		return rand.Intn(1000)
	default:
		panic(fmt.Sprintf("unsupported type %T", vtype))
	}
}

func Map[T, U any](s []T, f func(T) U) []U {
	result := make([]U, len(s))
	for i, v := range s {
		result[i] = f(v)
	}
	return result
}

func AsAny[T SupportedType](data []T) []any {
	return Map(data, func(v T) any {
		return v
	})
}

func Avg[T float64 | int](data []T) float64 {
	if len(data) == 0 {
		return 0
	}

	var sum float64
	for _, v := range data {
		sum += float64(v)
	}
	return sum / float64(len(data))
}
