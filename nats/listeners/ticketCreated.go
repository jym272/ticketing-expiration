package listeners

/*
{
  "tickets.created": {
    "id": 44417,
    "title": "New ticket",
    "price": 17
  }
}
*/

type TicketCreatedMessage struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	Price int    `json:"price"`
}

type Message struct {
	TicketsCreated TicketCreatedMessage `json:"tickets.created"`
}

//type Message struct {
//	TicketsCreated struct {
//		ID    int    `json:"id"`
//		Title string `json:"title"`
//		Price int    `json:"price"`
//	} `json:"tickets.created"`
//}
