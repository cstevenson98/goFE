package pokedex

type PokemonResultsList struct {
	Count   int              `json:"count"`
	Results []PokemonResults `json:"results"`
}

type PokemonResults struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}
