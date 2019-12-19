package domain

type CommandBody struct {
    Host     string   `json:"host"`
    User     string   `json:"user"`
    Password string   `json:"password"`
    Commands []string `json:"commands"`
}
