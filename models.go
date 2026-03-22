package main

type Student struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Programme string `json:"programme"`
	Year      int    `json:"year"`
}

type Course struct {
	ID       int64  `json:"id"`
	Code     string `json:"code"`
	Title    string `json:"title"`
	Credits  int    `json:"credits"`
	Enrolled int    `json:"enrolled"`
}
