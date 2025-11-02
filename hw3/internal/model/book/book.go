package book

type Book struct {
	id    int
	title string
	price float64
}

func NewBook(id int, title string, price float64) *Book {
	return &Book{
		id:    id,
		title: title,
		price: price,
	}
}

func (b *Book) GetID() int {
	return b.id
}

func (b *Book) GetTitle() string {
	return b.title
}

func (b *Book) GetPrice() float64 {
	return b.price
}
