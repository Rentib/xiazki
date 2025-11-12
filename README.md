# xiazki

Lightweight, selfhosted application for tracking books.

## Why?
I started reading books, and thought of tracking it. I saw goodreads and
I hated it. Then I remembered about self-hosting, so I looked for something
reasonable. There is [jelu](https://github.com/bayang/jelu), which would be
great if not for 3 things: it's slow, it's ugly and it uses A LOT of ram. There
is also [booklogr](https://github.com/Mozzo1000/booklogr), but it has even more
problems.

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
- [ ] color schemes
- [ ] importing books based on ISBN, author and title
- [ ] listing all books by author/tag/translator/narrator
- [x] listing books
- [ ] marking books as "to read"
- [ ] marking books as reading/finished/dropped
- [ ] merge authors
- [ ] quotes
- [ ] reading statistics
- [ ] reviews
- [ ] saving covers on disk
- [ ] sorting books
- [ ] track page
- [x] user accounts
- [ ] various editions of works
