package models

type EntryType string

var (
	EntryTypeNote EntryType = "note"
)

type Login struct {
	Username string `json:"username"`
	Password string `json:"password"`
	URL      string `json:"url"`
}

type Note struct {
	Text string `json:"text"`
}

type CreditCard struct {
	Number     string `json:"number"`
	Expiration string `json:"expiration"`
	CVV        string `json:"cvv"`
	Holder     string `json:"holder"`
}
