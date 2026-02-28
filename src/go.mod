module github.com/andrlikjira/dp-teals

replace github.com/andrlikjirka/merkle => ../pkg/merkle

replace github.com/andrlikjirka/hash => ../pkg/hash

replace github.com/andrlikjirka/mmr => ../pkg/mmr

go 1.25

require (
	github.com/andrlikjirka/merkle v0.0.0-00010101000000-000000000000
	github.com/andrlikjirka/mmr v0.0.0-00010101000000-000000000000
)

require github.com/andrlikjirka/hash v0.0.0-00010101000000-000000000000 // indirect
