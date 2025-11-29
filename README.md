# xiazki

Lightweight, selfhosted application for tracking books.

## Why?
I started reading books, and thought of tracking it. I saw goodreads and
I hated it. Then I remembered about self-hosting, so I looked for something
reasonable. There is [jelu](https://github.com/bayang/jelu), which would be
great if not for 3 things: it's slow, it's ugly, and it uses A LOT of ram. There
is also [booklogr](https://github.com/Mozzo1000/booklogr), but it has even more
problems. Currently, xiazki solves only the ram usage issue and none of the
other, however it also introduces its own problems :-).

## Tools it uses
- [echo](https://echo.labstack.com/)
- [templ](https://templ.guide/)
- [tailwind](https://tailwindcss.com/)
- [htmx](https://htmx.org/)
- [bun](https://bun.uptrace.dev/)
- [sqlite](https://pkg.go.dev/github.com/uptrace/bun/driver/sqliteshim)

## Features:
- [x] adding/deleting/editing books
- [x] book's information
- [x] importing books based on ISBN
- [x] listing all books by author
- [x] listing books
- [x] marking books as reading/finished/dropped
- [x] reviews (ratings)
- [x] user accounts
- [ ] color schemes
- [ ] importing books based on author and title
- [ ] marking books as "to read"
- [ ] merge authors
- [ ] quotes
- [ ] reading statistics
- [ ] reviews (opinions)
- [ ] saving covers on disk
- [ ] search feature to look up books with given properties
- [ ] sorting books
- [ ] track page
- [ ] translations
- [ ] various editions of works

## Requirements:
- go 1.25.4 (earlier versions probably work too)
- templ 0.3.960
- tailwind 4.1.17

## Building
```sh
make
```
