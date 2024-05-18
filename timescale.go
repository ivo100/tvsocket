package tvsocket

type HLOC struct {
	Time   int64   `json:"time"`
	Open   float64 `json:"open"`
	High   float64 `json:"high"`
	Low    float64 `json:"low"`
	Close  float64 `json:"close"`
	Volume int64   `json:"volume"`
}

type Timescale struct {
	P   []PElement `json:"p"`
	T   int64      `json:"t"`
	TMS int64      `json:"t_ms"`
	M   string     `json:"m"`
}

type PClass struct {
	Price Price `json:"price"`
	//Zoffset   *int64    `json:"zoffset,omitempty"`
	//IndexDiff []int64   `json:"index_diff,omitempty"`
	//Changes   []int64   `json:"changes,omitempty"`
	//Index     *int64    `json:"index,omitempty"`
	//Marks     [][]int64 `json:"marks,omitempty"`
}

type Price struct {
	Node string       `json:"node"`
	S    []IndexValue `json:"s"`
	T    string       `json:"t"`
	NS   NS           `json:"ns"`
	Lbs  Lbs          `json:"lbs"`
}

type Lbs struct {
	BarCloseTime int64 `json:"bar_close_time"`
}

type NS struct {
	D       string  `json:"d"`
	Indexes []int64 `json:"indexes"`
}

type IndexValue struct {
	I int64     `json:"i"`
	V []float64 `json:"v"`
}

type PElement struct {
	PClass *PClass
	String *string
}
