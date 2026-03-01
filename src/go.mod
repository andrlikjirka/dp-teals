module github.com/andrlikjira/dp-teals

replace github.com/andrlikjirka/logger => ../pkg/logger

replace github.com/andrlikjirka/merkle => ../pkg/merkle

replace github.com/andrlikjirka/hash => ../pkg/hash

replace github.com/andrlikjirka/mmr => ../pkg/mmr

go 1.25

require (
	github.com/andrlikjirka/logger v0.0.0-00010101000000-000000000000
	github.com/andrlikjirka/merkle v0.0.0-00010101000000-000000000000
	github.com/andrlikjirka/mmr v0.0.0-00010101000000-000000000000
	github.com/caarlos0/env/v10 v10.0.0
	github.com/go-playground/validator/v10 v10.30.1
	github.com/joho/godotenv v1.5.1
)

require (
	github.com/andrlikjirka/hash v0.0.0-00010101000000-000000000000 // indirect
	github.com/gabriel-vasile/mimetype v1.4.12 // indirect
	github.com/go-chi/chi/v5 v5.2.5 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/leodido/go-urn v1.4.0 // indirect
	golang.org/x/crypto v0.46.0 // indirect
	golang.org/x/sys v0.39.0 // indirect
	golang.org/x/text v0.32.0 // indirect
)
