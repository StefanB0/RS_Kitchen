package pkg

var Staff = []Cook{
	{
		Id:          1,
		Rank:        3,
		Proficiency: 4,
		Name:        "Gordon Ramsay",
		Catchphrase: "Hey, panini head, are you listening to me?",
	},
	{
		Id:          2,
		Rank:        2,
		Proficiency: 3,
		Name:        "Guy Fieri",
		Catchphrase: "Holy moly, Stromboli!",
	},
	{
		Id:          3,
		Rank:        2,
		Proficiency: 2,
		Name:        "Wang Zhao",
		Catchphrase: "Rice is power",
	},
	{
		Id:          4,
		Rank:        1,
		Proficiency: 2,
		Name:        "Vegan cook",
		Catchphrase: "Go eat some grass",
	},
}

var CookingAparatus = map[string]int{"oven": 2, "stove": 1}
