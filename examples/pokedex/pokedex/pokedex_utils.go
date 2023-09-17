package pokedex

import "strconv"

const baseURL = "https://pokeapi.co/api/v2/"

func ListPokemonURL(limit, offset int) string {
	return baseURL + "pokemon/?limit=" + strconv.Itoa(limit) + "&offset=" + strconv.Itoa(offset)
}

func PokemonURL(id int) string {
	return baseURL + "pokemon/" + string(rune(id))
}
