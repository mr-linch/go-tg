package tg

import (
	"encoding/json"
	"fmt"
)

// ReactionType it's type for describe content of Reaction.
// It can be one of:
//   - [ReactionTypeEmoji]
//   - [ReactionTypeCustomEmoji]
type ReactionType struct {
	Emoji       *ReactionTypeEmoji
	CustomEmoji *ReactionTypeCustomEmoji
	Paid        *ReactionTypePaid
}

// NewReactionTypeEmoji returns ReactionType with emoji subtype.
func NewReactionTypeEmoji(emoji string) ReactionType {
	return ReactionType{Emoji: &ReactionTypeEmoji{Emoji: emoji}}
}

// NewReactionTypeCustomEmoji returns ReactionType with custom emoji subtype.
func NewReactionTypeCustomEmoji(id string) ReactionType {
	return ReactionType{CustomEmoji: &ReactionTypeCustomEmoji{CustomEmojiID: id}}
}

func (reaction ReactionType) MarshalJSON() ([]byte, error) {
	switch {
	case reaction.Emoji != nil:
		reaction.Emoji.Type = "emoji"
		return json.Marshal(reaction.Emoji)
	case reaction.CustomEmoji != nil:
		reaction.CustomEmoji.Type = "custom_emoji"
		return json.Marshal(reaction.CustomEmoji)
	case reaction.Paid != nil:
		reaction.Paid.Type = "paid"
		return json.Marshal(reaction.Paid)
	default:
		return nil, fmt.Errorf("unknown ReactionType type")
	}
}

func (reaction *ReactionType) UnmarshalJSON(v []byte) error {
	var partial struct {
		Type string `json:"type"`
	}

	if err := json.Unmarshal(v, &partial); err != nil {
		return fmt.Errorf("unmarshal ReactionType partial: %w", err)
	}

	switch partial.Type {
	case "emoji":
		reaction.Emoji = &ReactionTypeEmoji{}
		return json.Unmarshal(v, reaction.Emoji)
	case "custom_emoji":
		reaction.CustomEmoji = &ReactionTypeCustomEmoji{}
		return json.Unmarshal(v, reaction.CustomEmoji)
	case "paid":
		reaction.Paid = &ReactionTypePaid{}
		return json.Unmarshal(v, reaction.Paid)
	default:
		fmt.Println(reaction)
		return fmt.Errorf("unknown ReactionType type: %s", partial.Type)
	}
}

func (reaction *ReactionType) Type() string {
	switch {
	case reaction.Emoji != nil:
		return "emoji"
	case reaction.CustomEmoji != nil:
		return "custom_emoji"
	default:
		return "unknown"
	}
}

// Ported from https://github.com/grammyjs/emoji/blob/main/src/reactiondata.ts

// Define all available reactions that can be used in the bot.
var (
	ReactionTypeEmojiThumbsUp                   = NewReactionTypeEmoji("👍")
	ReactionTypeEmojiThumbsDown                 = NewReactionTypeEmoji("👎")
	ReactionTypeEmojiRedHeart                   = NewReactionTypeEmoji("❤")
	ReactionTypeEmojiFire                       = NewReactionTypeEmoji("🔥")
	ReactionTypeEmojiSmilingFaceWithHearts      = NewReactionTypeEmoji("🥰")
	ReactionTypeEmojiClappingHands              = NewReactionTypeEmoji("👏")
	ReactionTypeEmojiBeamingFaceWithSmilingEyes = NewReactionTypeEmoji("😁")
	ReactionTypeEmojiThinkingFace               = NewReactionTypeEmoji("🤔")
	ReactionTypeEmojiExplodingHead              = NewReactionTypeEmoji("🤯")
	ReactionTypeEmojiFaceScreamingInFear        = NewReactionTypeEmoji("😱")
	ReactionTypeEmojiFaceWithSymbolsOnMouth     = NewReactionTypeEmoji("🤬")
	ReactionTypeEmojiCryingFace                 = NewReactionTypeEmoji("😢")
	ReactionTypeEmojiPartyPopper                = NewReactionTypeEmoji("🎉")
	ReactionTypeEmojiStarStruck                 = NewReactionTypeEmoji("🤩")
	ReactionTypeEmojiFaceVomiting               = NewReactionTypeEmoji("🤮")
	ReactionTypeEmojiPileOfPoo                  = NewReactionTypeEmoji("💩")
	ReactionTypeEmojiFoldedHands                = NewReactionTypeEmoji("🙏")
	ReactionTypeEmojiOkHand                     = NewReactionTypeEmoji("👌")
	ReactionTypeEmojiDove                       = NewReactionTypeEmoji("🕊")
	ReactionTypeEmojiClownFace                  = NewReactionTypeEmoji("🤡")
	ReactionTypeEmojiYawningFace                = NewReactionTypeEmoji("🥱")
	ReactionTypeEmojiWoozyFace                  = NewReactionTypeEmoji("🥴")
	ReactionTypeEmojiSmilingFaceWithHeartEyes   = NewReactionTypeEmoji("😍")
	ReactionTypeEmojiSpoutingWhale              = NewReactionTypeEmoji("🐳")
	ReactionTypeEmojiHeartOnFire                = NewReactionTypeEmoji("❤‍🔥")
	ReactionTypeEmojiNewMoonFace                = NewReactionTypeEmoji("🌚")
	ReactionTypeEmojiHotDog                     = NewReactionTypeEmoji("🌭")
	ReactionTypeEmojiHundredPoints              = NewReactionTypeEmoji("💯")
	ReactionTypeEmojiRollingOnTheFloorLaughing  = NewReactionTypeEmoji("🤣")
	ReactionTypeEmojiHighVoltage                = NewReactionTypeEmoji("⚡")
	ReactionTypeEmojiBanana                     = NewReactionTypeEmoji("🍌")
	ReactionTypeEmojiTrophy                     = NewReactionTypeEmoji("🏆")
	ReactionTypeEmojiBrokenHeart                = NewReactionTypeEmoji("💔")
	ReactionTypeEmojiFaceWithRaisedEyebrow      = NewReactionTypeEmoji("🤨")
	ReactionTypeEmojiNeutralFace                = NewReactionTypeEmoji("😐")
	ReactionTypeEmojiStrawberry                 = NewReactionTypeEmoji("🍓")
	ReactionTypeEmojiBottleWithPoppingCork      = NewReactionTypeEmoji("🍾")
	ReactionTypeEmojiKissMark                   = NewReactionTypeEmoji("💋")
	ReactionTypeEmojiMiddleFinger               = NewReactionTypeEmoji("🖕")
	ReactionTypeEmojiSmilingFaceWithHorns       = NewReactionTypeEmoji("😈")
	ReactionTypeEmojiSleepingFace               = NewReactionTypeEmoji("😴")
	ReactionTypeEmojiLoudlyCryingFace           = NewReactionTypeEmoji("😭")
	ReactionTypeEmojiNerdFace                   = NewReactionTypeEmoji("🤓")
	ReactionTypeEmojiGhost                      = NewReactionTypeEmoji("👻")
	ReactionTypeEmojiManTechnologist            = NewReactionTypeEmoji("👨‍💻")
	ReactionTypeEmojiEyes                       = NewReactionTypeEmoji("👀")
	ReactionTypeEmojiJackOLantern               = NewReactionTypeEmoji("🎃")
	ReactionTypeEmojiSeeNoEvilMonkey            = NewReactionTypeEmoji("🙈")
	ReactionTypeEmojiSmilingFaceWithHalo        = NewReactionTypeEmoji("😇")
	ReactionTypeEmojiFearfulFace                = NewReactionTypeEmoji("😨")
	ReactionTypeEmojiHandshake                  = NewReactionTypeEmoji("🤝")
	ReactionTypeEmojiWritingHand                = NewReactionTypeEmoji("✍")
	ReactionTypeEmojiSmilingFaceWithOpenHands   = NewReactionTypeEmoji("🤗")
	ReactionTypeEmojiSalutingFace               = NewReactionTypeEmoji("🫡")
	ReactionTypeEmojiSantaClaus                 = NewReactionTypeEmoji("🎅")
	ReactionTypeEmojiChristmasTree              = NewReactionTypeEmoji("🎄")
	ReactionTypeEmojiSnowman                    = NewReactionTypeEmoji("☃")
	ReactionTypeEmojiNailPolish                 = NewReactionTypeEmoji("💅")
	ReactionTypeEmojiZanyFace                   = NewReactionTypeEmoji("🤪")
	ReactionTypeEmojiMoai                       = NewReactionTypeEmoji("🗿")
	ReactionTypeEmojiCoolButton                 = NewReactionTypeEmoji("🆒")
	ReactionTypeEmojiHeartWithArrow             = NewReactionTypeEmoji("💘")
	ReactionTypeEmojiHearNoEvilMonkey           = NewReactionTypeEmoji("🙉")
	ReactionTypeEmojiUnicorn                    = NewReactionTypeEmoji("🦄")
	ReactionTypeEmojiFaceBlowingAKiss           = NewReactionTypeEmoji("😘")
	ReactionTypeEmojiPill                       = NewReactionTypeEmoji("💊")
	ReactionTypeEmojiSpeakNoEvilMonkey          = NewReactionTypeEmoji("🙊")
	ReactionTypeEmojiSmilingFaceWithSunglasses  = NewReactionTypeEmoji("😎")
	ReactionTypeEmojiAlienMonster               = NewReactionTypeEmoji("👾")
	ReactionTypeEmojiManShrugging               = NewReactionTypeEmoji("🤷‍♂")
	ReactionTypeEmojiPersonShrugging            = NewReactionTypeEmoji("🤷")
	ReactionTypeEmojiWomanShrugging             = NewReactionTypeEmoji("🤷‍♀")
	ReactionTypeEmojiEnragedFace                = NewReactionTypeEmoji("😡")
)

var (
	// ReactionTypeEmojiAll is a list of all available emoji reactions
	// that can be used in the bot as ReactionType.
	ReactionTypeEmojiAll = []ReactionType{
		ReactionTypeEmojiThumbsUp,
		ReactionTypeEmojiThumbsDown,
		ReactionTypeEmojiRedHeart,
		ReactionTypeEmojiFire,
		ReactionTypeEmojiSmilingFaceWithHearts,
		ReactionTypeEmojiClappingHands,
		ReactionTypeEmojiBeamingFaceWithSmilingEyes,
		ReactionTypeEmojiThinkingFace,
		ReactionTypeEmojiExplodingHead,
		ReactionTypeEmojiFaceScreamingInFear,
		ReactionTypeEmojiFaceWithSymbolsOnMouth,
		ReactionTypeEmojiCryingFace,
		ReactionTypeEmojiPartyPopper,
		ReactionTypeEmojiStarStruck,
		ReactionTypeEmojiFaceVomiting,
		ReactionTypeEmojiPileOfPoo,
		ReactionTypeEmojiFoldedHands,
		ReactionTypeEmojiOkHand,
		ReactionTypeEmojiDove,
		ReactionTypeEmojiClownFace,
		ReactionTypeEmojiYawningFace,
		ReactionTypeEmojiWoozyFace,
		ReactionTypeEmojiSmilingFaceWithHeartEyes,
		ReactionTypeEmojiSpoutingWhale,
		ReactionTypeEmojiHeartOnFire,
		ReactionTypeEmojiNewMoonFace,
		ReactionTypeEmojiHotDog,
		ReactionTypeEmojiHundredPoints,
		ReactionTypeEmojiRollingOnTheFloorLaughing,
		ReactionTypeEmojiHighVoltage,
		ReactionTypeEmojiBanana,
		ReactionTypeEmojiTrophy,
		ReactionTypeEmojiBrokenHeart,
		ReactionTypeEmojiFaceWithRaisedEyebrow,
		ReactionTypeEmojiNeutralFace,
		ReactionTypeEmojiStrawberry,
		ReactionTypeEmojiBottleWithPoppingCork,
		ReactionTypeEmojiKissMark,
		ReactionTypeEmojiMiddleFinger,
		ReactionTypeEmojiSmilingFaceWithHorns,
		ReactionTypeEmojiSleepingFace,
		ReactionTypeEmojiLoudlyCryingFace,
		ReactionTypeEmojiNerdFace,
		ReactionTypeEmojiGhost,
		ReactionTypeEmojiManTechnologist,
		ReactionTypeEmojiEyes,
		ReactionTypeEmojiJackOLantern,
		ReactionTypeEmojiSeeNoEvilMonkey,
		ReactionTypeEmojiSmilingFaceWithHalo,
		ReactionTypeEmojiFearfulFace,
		ReactionTypeEmojiHandshake,
		ReactionTypeEmojiWritingHand,
		ReactionTypeEmojiSmilingFaceWithOpenHands,
		ReactionTypeEmojiSalutingFace,
		ReactionTypeEmojiSantaClaus,
		ReactionTypeEmojiChristmasTree,
		ReactionTypeEmojiSnowman,
		ReactionTypeEmojiNailPolish,
		ReactionTypeEmojiZanyFace,
		ReactionTypeEmojiMoai,
		ReactionTypeEmojiCoolButton,
		ReactionTypeEmojiHeartWithArrow,
		ReactionTypeEmojiHearNoEvilMonkey,
		ReactionTypeEmojiUnicorn,
		ReactionTypeEmojiFaceBlowingAKiss,
		ReactionTypeEmojiPill,
		ReactionTypeEmojiSpeakNoEvilMonkey,
		ReactionTypeEmojiSmilingFaceWithSunglasses,
		ReactionTypeEmojiAlienMonster,
		ReactionTypeEmojiManShrugging,
		ReactionTypeEmojiPersonShrugging,
		ReactionTypeEmojiWomanShrugging,
		ReactionTypeEmojiEnragedFace,
	}
)
