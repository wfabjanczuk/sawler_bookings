# Fort Smythe Bed and Breakfast

Golang learning project. Simple web application for making room reservations in a fictional residence. Consists of a page for clients and admin panel for the property owner.

The repository is following Udemy course [Building Modern Web Applications with Go (Golang)](https://www.udemy.com/course/building-modern-web-applications-with-go) by professor [Trevor Sawler](https://github.com/tsawler).

## Development

Run: `docker-compose up`

When the application has started, make sure database migrations have been applied by executing: 

`docker exec -ti bookings_backend bash -c "soda migrate"`

By default the application is on `localhost:8080` and the database on `localhost:54321`

## Dependencies

- Built in Go version 1.17
- Backend uses:
  - [chi router](https://github.com/go-chi/chi)
  - [SCS: HTTP Session Management for Go](https://github.com/alexedwards/scs)
  - [nosurf](https://github.com/justinas/nosurf)
  - [govalidator](https://github.com/asaskevich/govalidator)
  - [pgx: PostgreSQL Driver and Toolkit](https://github.com/jackc/pgx)
  - [Go Simple Mail](https://github.com/xhit/go-simple-mail)
  - [GoDotEnv](https://github.com/joho/godotenv)
  - [Soda CLI](https://github.com/gobuffalo/pop)
  - [Go Compile Daemon](https://github.com/githubnemo/CompileDaemon)
- Frontend uses:
  - [Bootstrap 5](https://getbootstrap.com)
  - [Vanilla JS Datepicker](https://github.com/mymth/vanillajs-datepicker)
  - [notie](https://github.com/jaredreich/notie)
  - [sweetalert2](https://sweetalert2.github.io)
  - [RoyalUI Admin Template (Bootstrap 4)](https://github.com/BootstrapDash/RoyalUI-Free-Bootstrap-Admin-Template)
  - [Simple DataTables](https://github.com/fiduswriter/Simple-DataTables)
  - [Emoji Favicons](https://favicon.io/emoji-favicons)
