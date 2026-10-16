package builtin

import (
	"context"
	"strings"
	"sync"

	"github.com/avestura/zagros/internal/core"
)

type emoji struct {
	base
	once    sync.Once
	entries []emojiEntry
}

type emojiEntry struct{ glyph, name, hay string }

func newEmoji() *emoji { return &emoji{base: base{"emoji"}} }

func (e *emoji) Ranked() bool       { return true }
func (e *emoji) Keywords() []string { return []string{"emoji"} }

func (e *emoji) load() {
	for _, line := range strings.Split(emojiData, "\n") {
		p := strings.Split(strings.TrimSpace(line), "|")
		if len(p) < 2 {
			continue
		}
		hay := strings.ToLower(p[1])
		if len(p) > 2 {
			hay += " " + strings.ToLower(p[2])
		}
		e.entries = append(e.entries, emojiEntry{p[0], p[1], hay})
	}
}

func (e *emoji) Query(ctx context.Context, q string) []core.Item {
	var term string
	q = strings.TrimLeft(q, " ")
	if strings.HasPrefix(q, ":") {
		term = strings.TrimSpace(strings.TrimSuffix(q[1:], ":"))
	} else if rest, _, ok := trigger(q, "emoji"); ok {
		term = rest
	} else {
		return nil
	}
	e.once.Do(e.load)
	words := strings.Fields(strings.ToLower(term))
	var out []core.Item
	for _, en := range e.entries {
		match := true
		for _, w := range words {
			if !strings.Contains(en.hay, w) {
				match = false
				break
			}
		}
		if !match {
			continue
		}
		out = append(out, core.Item{
			ID: "emoji:" + en.name, Title: en.glyph + "  " + en.name, Subtitle: "Copy emoji", Category: "Emoji",
			Action: copyAction(en.glyph),
		})
		if len(out) >= 24 {
			break
		}
	}
	// prefer entries whose name starts with the first word
	if len(words) > 0 {
		pri := out[:0:0]
		rest := out[:0:0]
		for _, it := range out {
			n := strings.ToLower(strings.TrimSpace(strings.SplitN(it.Title, "  ", 2)[1]))
			if strings.HasPrefix(n, words[0]) {
				pri = append(pri, it)
			} else {
				rest = append(rest, it)
			}
		}
		out = append(pri, rest...)
	}
	return out
}

// emojiData: glyph|name|extra keywords
const emojiData = `😀|grinning face|smile happy
😃|smiley|happy joy
😄|smile|happy laugh
😁|beaming face|grin
😆|laughing|lol haha
😅|sweat smile|nervous relief
🤣|rofl|rolling floor laughing lol
😂|joy|tears laugh lol
🙂|slight smile|ok
🙃|upside down|silly sarcasm
😉|wink|flirt
😊|blush|happy shy
😇|innocent|angel halo
🥰|smiling face with hearts|love adore
😍|heart eyes|love
🤩|star struck|wow amazing
😘|kiss|blow kiss love
😋|yum|tasty delicious
😛|tongue|playful
😜|wink tongue|crazy joke
🤪|zany|crazy wild
🤗|hug|hugging
🤔|thinking|hmm think
🤨|raised eyebrow|skeptical
😐|neutral|meh
😑|expressionless|blank
😶|no mouth|silent speechless
🙄|eye roll|annoyed whatever
😏|smirk|smug
😒|unamused|bored
😬|grimace|awkward
😌|relieved|calm peace
😔|pensive|sad thoughtful
😪|sleepy|tired
😴|sleeping|zzz sleep
😷|mask|sick medical
🤒|thermometer face|ill fever
🤕|bandage|hurt injured
🤢|nauseated|sick green
🤮|vomit|sick throw up
🥵|hot face|heat sweating
🥶|cold face|freezing
😵|dizzy face|knocked out
🤯|mind blown|exploding head shocked
🤠|cowboy|hat yeehaw
🥳|party|celebrate birthday
😎|sunglasses|cool
🤓|nerd|geek glasses
🧐|monocle|inspect curious
😕|confused|puzzled
😟|worried|anxious
🙁|frown|sad
😮|open mouth|surprised wow
😲|astonished|shocked
😳|flushed|embarrassed
🥺|pleading|puppy eyes please
😦|frowning|dismay
😨|fearful|scared
😰|anxious sweat|nervous
😢|cry|sad tear
😭|sob|crying loud
😱|scream|fear horror
😖|confounded|frustrated
😞|disappointed|sad
😓|downcast sweat|hard work
😩|weary|tired
😫|tired face|exhausted
🥱|yawn|bored sleepy
😤|triumph|huff angry steam
😡|rage|angry mad red
😠|angry|mad
🤬|cursing|swearing symbols
😈|smiling devil|evil imp
👿|angry devil|evil
💀|skull|dead death
☠️|skull crossbones|danger pirate
💩|poop|poo
🤡|clown|joker
👻|ghost|halloween boo
👽|alien|ufo
🤖|robot|bot ai
🎃|pumpkin|halloween
😺|cat smile|happy cat
😸|cat grin|happy cat
😹|cat joy|laughing cat
😻|cat heart eyes|love cat
🙈|see no evil|monkey hide
🙉|hear no evil|monkey
🙊|speak no evil|monkey secret
💋|kiss mark|lips
👋|wave|hello hi bye
🤚|raised back of hand|stop
✋|raised hand|high five stop
🖖|vulcan|spock star trek
👌|ok hand|perfect
🤌|pinched fingers|italian
✌️|victory|peace
🤞|crossed fingers|luck hope
🤟|love you gesture|rock
🤘|rock on|metal horns
🤙|call me|shaka
👈|point left|left
👉|point right|right
👆|point up|up
👇|point down|down
☝️|index up|one
👍|thumbs up|like yes approve +1
👎|thumbs down|dislike no -1
✊|fist|power punch
👊|punch|fist bump
👏|clap|applause bravo
🙌|raised hands|celebrate hooray
👐|open hands|jazz hands
🤝|handshake|deal agreement
🙏|pray|please thanks namaste hope
✍️|writing hand|write
💪|muscle|strong flex biceps
🦾|mechanical arm|robot prosthetic
👀|eyes|look see watching
👁️|eye|look
👅|tongue out|taste
👄|mouth|lips
🧠|brain|smart mind
🫀|heart organ|anatomical
👶|baby|infant child
🧑|person|adult
👨|man|male
👩|woman|female
🧔|beard|man
👴|old man|grandfather
👵|old woman|grandmother
👨‍💻|man technologist|developer coder programmer
👩‍💻|woman technologist|developer coder programmer
🕵️|detective|spy investigate
🧙|mage|wizard magic
🧛|vampire|dracula
🧟|zombie|undead
🧞|genie|wish
🧜|merperson|mermaid
🧚|fairy|magic
🦸|superhero|hero
🦹|supervillain|villain
💃|dancer|dance
🕺|man dancing|dance
🏃|runner|running
🚶|walking|walk
🧘|lotus position|yoga meditate
❤️|red heart|love
🧡|orange heart|love
💛|yellow heart|love
💚|green heart|love
💙|blue heart|love
💜|purple heart|love
🖤|black heart|love dark
🤍|white heart|love
🤎|brown heart|love
💔|broken heart|sad heartbreak
❣️|heart exclamation|love
💕|two hearts|love
💞|revolving hearts|love
💓|beating heart|love
💗|growing heart|love
💖|sparkling heart|love
💘|heart arrow|cupid love
💝|heart ribbon|gift love
💯|hundred|100 perfect score
💢|anger|symbol
💥|collision|boom bang
💫|dizzy|star
💦|sweat drops|water splash
💨|dash|wind fast
💬|speech bubble|chat message
💭|thought bubble|thinking
💤|zzz|sleep
🕳️|hole|pit
🐶|dog|puppy pet
🐱|cat|kitten pet
🐭|mouse|rodent
🐹|hamster|pet
🐰|rabbit|bunny
🦊|fox|animal
🐻|bear|animal
🐼|panda|animal
🐨|koala|animal
🐯|tiger|animal
🦁|lion|animal king
🐮|cow|animal
🐷|pig|animal
🐸|frog|animal
🐵|monkey|animal
🐔|chicken|animal
🐧|penguin|animal linux
🐦|bird|animal
🐤|chick|baby bird
🦆|duck|animal
🦅|eagle|bird
🦉|owl|bird
🦇|bat|animal
🐺|wolf|animal
🐴|horse|animal
🦄|unicorn|magic
🐝|bee|insect
🐛|bug|insect caterpillar
🦋|butterfly|insect
🐌|snail|slow
🐞|lady beetle|ladybug
🐜|ant|insect
🪲|beetle|bug
🕷️|spider|arachnid
🐢|turtle|slow
🐍|snake|python reptile
🦎|lizard|reptile
🦖|t-rex|dinosaur
🐙|octopus|sea
🦑|squid|sea
🦀|crab|rust sea
🐠|tropical fish|sea
🐟|fish|sea
🐬|dolphin|sea
🐳|whale|sea docker
🦈|shark|sea
🐊|crocodile|reptile
🐘|elephant|animal
🦒|giraffe|animal
🐪|camel|desert
🐑|sheep|animal
🐐|goat|animal
🌵|cactus|desert plant
🎄|christmas tree|holiday
🌲|evergreen|tree
🌳|tree|deciduous
🌴|palm tree|tropical
🌱|seedling|plant sprout
🌿|herb|plant
🍀|four leaf clover|luck
🍁|maple leaf|canada autumn
🍂|fallen leaf|autumn
🌸|cherry blossom|flower spring
🌹|rose|flower love
🌻|sunflower|flower
🌷|tulip|flower
💐|bouquet|flowers
🍄|mushroom|fungus
🌍|globe europe africa|earth world
🌎|globe americas|earth world
🌏|globe asia|earth world
🌙|crescent moon|night
🌕|full moon|night
⭐|star|favorite
🌟|glowing star|sparkle
✨|sparkles|magic new shine
⚡|high voltage|lightning zap power
🔥|fire|hot lit flame
🌈|rainbow|pride
☀️|sun|sunny weather
⛅|sun behind cloud|weather
☁️|cloud|weather
🌧️|rain|weather
⛈️|thunderstorm|weather
❄️|snowflake|cold winter
☃️|snowman|winter
🌊|wave|ocean sea
💧|droplet|water
🍏|green apple|fruit
🍎|red apple|fruit
🍐|pear|fruit
🍊|orange|tangerine fruit
🍋|lemon|fruit
🍌|banana|fruit
🍉|watermelon|fruit
🍇|grapes|fruit
🍓|strawberry|fruit
🫐|blueberries|fruit
🍒|cherries|fruit
🍑|peach|fruit
🥭|mango|fruit
🍍|pineapple|fruit
🥥|coconut|fruit
🥝|kiwi|fruit
🍅|tomato|vegetable
🥑|avocado|fruit
🥕|carrot|vegetable
🌽|corn|vegetable
🌶️|hot pepper|spicy chili
🥔|potato|vegetable
🍞|bread|food
🥐|croissant|food
🧀|cheese|food
🥚|egg|food
🍳|cooking|frying egg
🥓|bacon|food
🍔|hamburger|burger food
🍟|fries|food
🍕|pizza|food
🌭|hot dog|food
🌮|taco|food
🌯|burrito|food
🥗|salad|food healthy
🍝|spaghetti|pasta food
🍜|ramen|noodles food
🍣|sushi|food
🍤|fried shrimp|food
🍦|ice cream|dessert
🍩|doughnut|dessert donut
🍪|cookie|dessert
🎂|birthday cake|dessert party
🍰|cake|dessert
🍫|chocolate|dessert
🍬|candy|sweet
🍿|popcorn|movie snack
☕|coffee|hot drink tea
🍵|tea|green tea drink
🍺|beer|drink
🍻|beers|cheers drink
🥂|clinking glasses|cheers champagne
🍷|wine|drink
🍸|cocktail|drink
🥤|cup with straw|soda drink
⚽|soccer|football sport
🏀|basketball|sport
🏈|american football|sport
⚾|baseball|sport
🎾|tennis|sport
🏐|volleyball|sport
🎱|pool 8 ball|billiards
🏓|ping pong|table tennis
🥊|boxing glove|sport
🎯|direct hit|target dart goal bullseye
🎮|video game|controller gaming
🕹️|joystick|gaming
🎲|dice|game random
♟️|chess pawn|game
🎨|artist palette|art paint
🎬|clapper board|movie film
🎤|microphone|sing karaoke
🎧|headphones|music
🎵|musical note|music
🎶|musical notes|music
🎹|piano|music keyboard
🎸|guitar|music
🥁|drum|music
🎻|violin|music
🏆|trophy|win award champion
🥇|gold medal|first winner
🥈|silver medal|second
🥉|bronze medal|third
🎁|gift|present
🎉|party popper|celebrate tada congratulations
🎊|confetti ball|celebrate
🎈|balloon|party
🚗|car|automobile
🚕|taxi|cab
🚌|bus|transport
🚑|ambulance|emergency
🚒|fire engine|emergency
🚓|police car|emergency
🚲|bicycle|bike
🛴|kick scooter|transport
🏍️|motorcycle|bike
🚂|locomotive|train
🚄|bullet train|fast train
✈️|airplane|flight travel
🚀|rocket|launch ship fast space
🛸|flying saucer|ufo
🚁|helicopter|transport
⛵|sailboat|boat
🚢|ship|boat cruise
⚓|anchor|boat
🗺️|world map|travel
🗽|statue of liberty|new york
🗼|tokyo tower|japan
🏰|castle|europe
🏠|house|home
🏢|office building|work
🏥|hospital|health
🏦|bank|money
🏫|school|education
⛪|church|religion
🕌|mosque|religion
🕍|synagogue|religion
⌚|watch|time
📱|mobile phone|smartphone
💻|laptop|computer
🖥️|desktop|computer pc
⌨️|keyboard|typing
🖱️|computer mouse|pc
💾|floppy disk|save
💿|cd|disc
📷|camera|photo
📹|video camera|film
📺|television|tv
📻|radio|music
🔋|battery|power
🔌|electric plug|power
💡|light bulb|idea
🔦|flashlight|torch
🕯️|candle|light
💰|money bag|rich dollar
💵|dollar|cash money
💳|credit card|payment
💎|gem|diamond jewel
⚖️|balance scale|justice law
🔧|wrench|tool fix
🔨|hammer|tool build
🛠️|hammer and wrench|tools settings
⚙️|gear|settings cog
🧰|toolbox|tools
🔩|nut and bolt|hardware
🔗|link|chain url
⛓️|chains|link
🧲|magnet|attract
🔒|locked|lock security private
🔓|unlocked|open security
🔑|key|password access
🗝️|old key|secret
🛡️|shield|security protect
🔍|magnifying glass left|search find
🔎|magnifying glass right|search find zoom
📧|e-mail|email mail message
📨|incoming envelope|email
📦|package|box parcel delivery
📝|memo|note write todo
📁|file folder|directory
📂|open folder|directory
📅|calendar|date schedule
📆|tear-off calendar|date
📊|bar chart|stats graph
📈|chart increasing|growth up
📉|chart decreasing|down
📋|clipboard|paste list
📌|pushpin|pin
📍|round pushpin|location pin
📎|paperclip|attach
✂️|scissors|cut
🗑️|wastebasket|trash delete bin
📚|books|library read
📖|open book|read
🔖|bookmark|save
🏷️|label|tag
📢|loudspeaker|announce
🔔|bell|notification alert
🔕|bell slash|mute silent
⏰|alarm clock|time wake
⏳|hourglass|time waiting
⌛|hourglass done|time
✅|check mark button|done yes ok
☑️|check box|done
✔️|check mark|done yes
❌|cross mark|no wrong delete
❎|cross button|no
⭕|hollow red circle|circle
🚫|prohibited|forbidden no
⛔|no entry|stop
⚠️|warning|caution alert
❗|exclamation|important
❓|question|help
❔|white question|help
‼️|double exclamation|important
➕|plus|add
➖|minus|subtract
➗|divide|division
✖️|multiply|times
♾️|infinity|forever
➡️|right arrow|next
⬅️|left arrow|back previous
⬆️|up arrow|up
⬇️|down arrow|down
↩️|return arrow|undo
🔄|counterclockwise arrows|refresh reload sync
🔁|repeat|loop
🔀|shuffle|random
▶️|play|start
⏸️|pause|wait
⏹️|stop|end
⏪|rewind|back
⏩|fast forward|skip
🔴|red circle|record
🟠|orange circle|circle
🟡|yellow circle|circle
🟢|green circle|online ok
🔵|blue circle|circle
🟣|purple circle|circle
⚫|black circle|circle
⚪|white circle|circle
🟥|red square|square
🟩|green square|square
🟦|blue square|square
🏁|chequered flag|finish race
🚩|triangular flag|flag
🏳️‍🌈|rainbow flag|pride
🇺🇸|flag united states|usa america
🇬🇧|flag united kingdom|uk britain
🇨🇦|flag canada|canada
🇩🇪|flag germany|germany
🇫🇷|flag france|france
🇮🇹|flag italy|italy
🇪🇸|flag spain|spain
🇯🇵|flag japan|japan
🇨🇳|flag china|china
🇮🇳|flag india|india
🇧🇷|flag brazil|brazil
🇮🇷|flag iran|iran persia
🇹🇷|flag turkey|turkey
🇺🇦|flag ukraine|ukraine
©️|copyright|c
®️|registered|r
™️|trade mark|tm
§|section sign|legal
°|degree|temperature
µ|micro|mu
€|euro|currency
£|pound|currency
¥|yen|currency
₿|bitcoin|crypto currency
→|arrow right|rightwards
←|arrow left|leftwards
↑|arrow up|upwards
↓|arrow down|downwards
≈|approximately equal|almost
≠|not equal|different
≤|less or equal|math
≥|greater or equal|math
±|plus minus|math
√|square root|math
∞|infinity symbol|math
π|pi|math
✓|check|tick
✗|ballot x|cross
…|ellipsis|dots
•|bullet|dot
·|middle dot|interpunct
¯\_(ツ)_/¯|shrug|whatever dunno
(╯°□°)╯︵ ┻━┻|table flip|angry
┬─┬ノ( º _ ºノ)|table unflip|calm
ಠ_ಠ|disapproval|look
( ͡° ͜ʖ ͡°)|lenny|face`
