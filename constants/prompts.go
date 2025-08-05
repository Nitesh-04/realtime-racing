package constants

import (
	"math/rand"
)

type Prompt struct {
	ID     int    `json:"id"`
	Prompt string `json:"prompt"`
}

var Prompts = []Prompt{
	{
		ID: 1,
		Prompt: "the ancient stone tablet was covered in a language no one could decipher intricate carvings depicted constellations and mythical beasts hinting at a lost civilizations knowledge of the cosmos an archeologist carefully dusted the surface revealing a single glowing symbol that pulsed with a faint ethereal light",
	},
	{
		ID: 2,
		Prompt: "a lone spaceship drifted through the silent star filled void its engine a marvel of intergalactic engineering had failed the pilot a grizzled veteran of countless deep space missions initiated a desperate reboot sequence hoping to restore power before the ships life support systems gave out completely",
	},
	{
		ID: 3,
		Prompt: "she carefully mixed the potions in her cauldron each ingredient was added with precision a pinch of moonflower a drop of captured starlight and a whispered incantation the liquid inside shimmered with an otherworldly glow promising to grant her the power to see into the future but at a terrible price",
	},
	{
		ID: 4,
		Prompt: "the detective stared at the cryptic note left at the scene of the crime the handwriting was elegant almost artistic but the message itself was a puzzle of riddles and metaphors he knew the solution lay hidden within the poetry a secret message revealing the next move of a brilliant criminal mastermind",
	},
	{
		ID: 5,
		Prompt: "a small robotic companion scurried through the overgrown ruins of a forgotten city its primary directive was to find and retrieve a lost data chip but its sensors detected something far more significant a faint signal emanated from deep within the rubble a message from an ancient long vanished artificial intelligence",
	},
	{
		ID: 6,
		Prompt: "the old lighthouse keeper polished the lens a ritual he had performed for decades a storm was brewing on the horizon and the lights beam was the only beacon of hope for ships caught in the tumultuous waves he watched the churning sea a testament to natures untamable power",
	},
	{
		ID: 7,
		Prompt: "she opened the antique music box and a delicate haunting melody filled the room the tiny ballerina inside pirouetted gracefully a silent dance that evoked memories of a childhood long past the music box a gift from her grandmother held a secret compartment she had never discovered before",
	},
	{
		ID: 8,
		Prompt: "the master thief crept through the museums security laser grid his movements were fluid and silent a ghost in the night his target was a priceless diamond a relic of a forgotten kingdom every step was a calculated risk one wrong move and the silent alarm would trigger",
	},
	{
		ID: 9,
		Prompt: "the wizards apprentice practiced his first spell a simple levitation charm with a flick of his wrist and a whispered word of power a small feather rose from the table hovering precariously in the air it was a clumsy attempt but a testament to his burgeoning magical abilities",
	},
	{
		ID: 10,
		Prompt: "a group of explorers trekked through the dense uncharted jungle a map drawn on brittle parchment guided their way to a legendary lost temple the air was thick with the sounds of exotic wildlife and every rustle of leaves could signal a hidden danger they pushed forward driven by curiosity and the promise of discovery",
	},
	{
		ID: 11,
		Prompt: "the sentient robot painter carefully selected its colors a canvas awaited blank and full of possibility the artist program a complex algorithm of aesthetics and emotion guided its metallic hand it began to paint creating a masterpiece that expressed a machines unique interpretation of human feelings and the natural world",
	},
	{
		ID: 12,
		Prompt: "he found the forgotten diary in the attic its pages yellowed and brittle with age the ink was faded but the words told a gripping tale of adventure betrayal and a hidden treasure he realized the diarys author was his great grandfather and the treasure was still waiting to be found",
	},
	{
		ID: 13,
		Prompt: "the captain of the airship surveyed the clouds below they formed a sea of white a vast endless landscape in the sky his crew were busy with their duties preparing for a long journey to a floating city the airships steam engines hummed with a powerful rhythmic beat a constant reminder of their upward momentum",
	},
	{
		ID: 14,
		Prompt: "a mysterious portal shimmered in the center of the forest it pulsed with a soft inviting light hinting at another world beyond its swirling surface a young adventurer driven by a thirst for the unknown took a deep breath and stepped through ready to face whatever lay on the other side",
	},
	{
		ID: 15,
		Prompt: "the dragon a magnificent beast of scales and fire rested atop a mountain peak its hoard of gold and jewels glittered in the afternoon sun a kings ransom but the dragon was not a creature of greed it was the sworn protector of the mountain a silent guardian of an ancient prophecy",
	},
	{
		ID: 16,
		Prompt: "the young alchemist finally created the philosophers stone it wasnt a glittering gem but a simple polished stone that radiated a quiet warmth he held it in his hand a tangible symbol of his years of tireless study and experimentation the stone could turn lead into gold but its true power was far more profound",
	},
	{
		ID: 17,
		Prompt: "the cybernetic warrior stood on the battlefield the last line of defense against an invading alien force his titanium armor was dented and scarred a testament to past battles but his optical sensors remained sharp and his internal processors were calculating the most effective strategy to defeat the enemy",
	},
	{
		ID: 18,
		Prompt: "she discovered a hidden garden behind a crumbling brick wall the plants inside were unlike anything she had ever seen with glowing petals and leaves that changed color with the passing of the hours it was a place of magic and tranquility a secret sanctuary she had stumbled upon",
	},
	{
		ID: 19,
		Prompt: "the legendary ghost ship the silent whisper sailed through the misty sea its sails were tattered and its wooden hull was rotting but it moved with an eerie grace its crew of spectral sailors going about their eternal duties it was a ship of lost souls forever bound to the vast ocean",
	},
	{
		ID: 20,
		Prompt: "the time traveler adjusted the dials on his machine a faint whirring sound filled the room as the temporal displacement engine began to power up he was about to embark on his most dangerous mission yet a journey to the distant past to witness a historical event and change the course of human history",
	},
}

// returns a random prompt from the list of prompts

func GetRandomPrompt() string {
	randomIndex := rand.Intn(len(Prompts))
	return Prompts[randomIndex].Prompt
}