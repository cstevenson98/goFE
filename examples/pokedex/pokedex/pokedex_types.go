package pokedex

type PokemonResultsList struct {
	Count   int              `json:"count"`
	Results []PokemonResults `json:"results"`
}

type PokemonResults struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	URL  string `json:"url"`
}
