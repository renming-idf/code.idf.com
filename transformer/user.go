package transformer

type User struct {
	ID         uint   `json:"id"`
	Phone      string `json:"phone"`
	Name       string `json:"name"`
	PublicKey  string `json:"public_key"`
	PrivateKey string `json:"private_key"`
	ParentID   uint   `json:"parent_id"`
}
