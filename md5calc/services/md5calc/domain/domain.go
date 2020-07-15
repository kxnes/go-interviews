package domain

type Status string

const (
	NotExist Status = "not exist"
	Pending  Status = "pending"
	Success  Status = "success"
	Failure  Status = "failure"
)

type Path struct {
	URL      string
	Query    string
	Status   Status
	Checksum Checksum
}

type Checksum struct {
	Sum string
}
