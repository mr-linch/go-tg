package tg

func newEmojiReaction(emoji string) ReactionType {
	return ReactionType{Emoji: &ReactionTypeEmoji{Emoji: emoji}}
}

func newCustomEmojiReaction(id string) ReactionType {
	return ReactionType{CustomEmoji: &ReactionTypeCustomEmoji{CustomEmojiID: id}}
}

// Ported from https://github.com/grammyjs/emoji/blob/main/src/reactiondata.ts

// Define all available reactions that can be used in the bot.
var (
	ReactionTypeEmojiThumbsUp                   = newEmojiReaction("👍")
	ReactionTypeEmojiThumbsDown                 = newEmojiReaction("👎")
	ReactionTypeEmojiRedHeart                   = newEmojiReaction("❤")
	ReactionTypeEmojiFire                       = newEmojiReaction("🔥")
	ReactionTypeEmojiSmilingFaceWithHearts      = newEmojiReaction("🥰")
	ReactionTypeEmojiClappingHands              = newEmojiReaction("👏")
	ReactionTypeEmojiBeamingFaceWithSmilingEyes = newEmojiReaction("😁")
	ReactionTypeEmojiThinkingFace               = newEmojiReaction("🤔")
	ReactionTypeEmojiExplodingHead              = newEmojiReaction("🤯")
	ReactionTypeEmojiFaceScreamingInFear        = newEmojiReaction("😱")
	ReactionTypeEmojiFaceWithSymbolsOnMouth     = newEmojiReaction("🤬")
	ReactionTypeEmojiCryingFace                 = newEmojiReaction("😢")
	ReactionTypeEmojiPartyPopper                = newEmojiReaction("🎉")
	ReactionTypeEmojiStarStruck                 = newEmojiReaction("🤩")
	ReactionTypeEmojiFaceVomiting               = newEmojiReaction("🤮")
	ReactionTypeEmojiPileOfPoo                  = newEmojiReaction("💩")
	ReactionTypeEmojiFoldedHands                = newEmojiReaction("🙏")
	ReactionTypeEmojiOkHand                     = newEmojiReaction("👌")
	ReactionTypeEmojiDove                       = newEmojiReaction("🕊")
	ReactionTypeEmojiClownFace                  = newEmojiReaction("🤡")
	ReactionTypeEmojiYawningFace                = newEmojiReaction("🥱")
	ReactionTypeEmojiWoozyFace                  = newEmojiReaction("🥴")
	ReactionTypeEmojiSmilingFaceWithHeartEyes   = newEmojiReaction("😍")
	ReactionTypeEmojiSpoutingWhale              = newEmojiReaction("🐳")
	ReactionTypeEmojiHeartOnFire                = newEmojiReaction("❤‍🔥")
	ReactionTypeEmojiNewMoonFace                = newEmojiReaction("🌚")
	ReactionTypeEmojiHotDog                     = newEmojiReaction("🌭")
	ReactionTypeEmojiHundredPoints              = newEmojiReaction("💯")
	ReactionTypeEmojiRollingOnTheFloorLaughing  = newEmojiReaction("🤣")
	ReactionTypeEmojiHighVoltage                = newEmojiReaction("⚡")
	ReactionTypeEmojiBanana                     = newEmojiReaction("🍌")
	ReactionTypeEmojiTrophy                     = newEmojiReaction("🏆")
	ReactionTypeEmojiBrokenHeart                = newEmojiReaction("💔")
	ReactionTypeEmojiFaceWithRaisedEyebrow      = newEmojiReaction("🤨")
	ReactionTypeEmojiNeutralFace                = newEmojiReaction("😐")
	ReactionTypeEmojiStrawberry                 = newEmojiReaction("🍓")
	ReactionTypeEmojiBottleWithPoppingCork      = newEmojiReaction("🍾")
	ReactionTypeEmojiKissMark                   = newEmojiReaction("💋")
	ReactionTypeEmojiMiddleFinger               = newEmojiReaction("🖕")
	ReactionTypeEmojiSmilingFaceWithHorns       = newEmojiReaction("😈")
	ReactionTypeEmojiSleepingFace               = newEmojiReaction("😴")
	ReactionTypeEmojiLoudlyCryingFace           = newEmojiReaction("😭")
	ReactionTypeEmojiNerdFace                   = newEmojiReaction("🤓")
	ReactionTypeEmojiGhost                      = newEmojiReaction("👻")
	ReactionTypeEmojiManTechnologist            = newEmojiReaction("👨‍💻")
	ReactionTypeEmojiEyes                       = newEmojiReaction("👀")
	ReactionTypeEmojiJackOLantern               = newEmojiReaction("🎃")
	ReactionTypeEmojiSeeNoEvilMonkey            = newEmojiReaction("🙈")
	ReactionTypeEmojiSmilingFaceWithHalo        = newEmojiReaction("😇")
	ReactionTypeEmojiFearfulFace                = newEmojiReaction("😨")
	ReactionTypeEmojiHandshake                  = newEmojiReaction("🤝")
	ReactionTypeEmojiWritingHand                = newEmojiReaction("✍")
	ReactionTypeEmojiSmilingFaceWithOpenHands   = newEmojiReaction("🤗")
	ReactionTypeEmojiSalutingFace               = newEmojiReaction("🫡")
	ReactionTypeEmojiSantaClaus                 = newEmojiReaction("🎅")
	ReactionTypeEmojiChristmasTree              = newEmojiReaction("🎄")
	ReactionTypeEmojiSnowman                    = newEmojiReaction("☃")
	ReactionTypeEmojiNailPolish                 = newEmojiReaction("💅")
	ReactionTypeEmojiZanyFace                   = newEmojiReaction("🤪")
	ReactionTypeEmojiMoai                       = newEmojiReaction("🗿")
	ReactionTypeEmojiCoolButton                 = newEmojiReaction("🆒")
	ReactionTypeEmojiHeartWithArrow             = newEmojiReaction("💘")
	ReactionTypeEmojiHearNoEvilMonkey           = newEmojiReaction("🙉")
	ReactionTypeEmojiUnicorn                    = newEmojiReaction("🦄")
	ReactionTypeEmojiFaceBlowingAKiss           = newEmojiReaction("😘")
	ReactionTypeEmojiPill                       = newEmojiReaction("💊")
	ReactionTypeEmojiSpeakNoEvilMonkey          = newEmojiReaction("🙊")
	ReactionTypeEmojiSmilingFaceWithSunglasses  = newEmojiReaction("😎")
	ReactionTypeEmojiAlienMonster               = newEmojiReaction("👾")
	ReactionTypeEmojiManShrugging               = newEmojiReaction("🤷‍♂")
	ReactionTypeEmojiPersonShrugging            = newEmojiReaction("🤷")
	ReactionTypeEmojiWomanShrugging             = newEmojiReaction("🤷‍♀")
	ReactionTypeEmojiEnragedFace                = newEmojiReaction("😡")
)

// ReactionTypeEmojiAll is a list of all available emoji reactions
// that can be used in the bot as ReactionType.
var ReactionTypeEmojiAll = []ReactionType{
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
