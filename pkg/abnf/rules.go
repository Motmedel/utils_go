package abnf

// Rule names of the ABNF grammar of ABNF (RFC 5234 Section 4, with the
// RFC 7405 char-val extension) and the core rules (RFC 5234 Section 8.1).
const (
	ruleNameRulelist              = "rulelist"
	ruleNameRule                  = "rule"
	ruleNameRulename              = "rulename"
	ruleNameDefinedAs             = "defined-as"
	ruleNameElements              = "elements"
	ruleNameCWsp                  = "c-wsp"
	ruleNameCNl                   = "c-nl"
	ruleNameComment               = "comment"
	ruleNameAlternation           = "alternation"
	ruleNameConcatenation         = "concatenation"
	ruleNameRepetition            = "repetition"
	ruleNameRepeat                = "repeat"
	ruleNameElement               = "element"
	ruleNameGroup                 = "group"
	ruleNameOption                = "option"
	ruleNameCharVal               = "char-val"
	ruleNameCaseInsensitiveString = "case-insensitive-string"
	ruleNameCaseSensitiveString   = "case-sensitive-string"
	ruleNameQuotedString          = "quoted-string"
	ruleNameNumVal                = "num-val"
	ruleNameBinVal                = "bin-val"
	ruleNameDecVal                = "dec-val"
	ruleNameHexVal                = "hex-val"
	ruleNameProseVal              = "prose-val"
	ruleNameAlpha                 = "ALPHA"
	ruleNameBit                   = "BIT"
	ruleNameChar                  = "CHAR"
	ruleNameCr                    = "CR"
	ruleNameCrlf                  = "CRLF"
	ruleNameCtl                   = "CTL"
	ruleNameDigit                 = "DIGIT"
	ruleNameDquote                = "DQUOTE"
	ruleNameHexdig                = "HEXDIG"
	ruleNameHtab                  = "HTAB"
	ruleNameLf                    = "LF"
	ruleNameLwsp                  = "LWSP"
	ruleNameOctet                 = "OCTET"
	ruleNameSp                    = "SP"
	ruleNameVchar                 = "VCHAR"
	ruleNameWsp                   = "WSP"
)

// Core rules (RFC 5234 Section 8.1).
var (
	// alpha is the rule `ALPHA = %x41-5A / %x61-7A`.
	alpha = newRule(ruleNameAlpha, cat(one(numRange("x", "41", "5A"))), cat(one(numRange("x", "61", "7A"))))

	// bit is the rule `BIT = "0" / "1"`.
	bit = newRule(ruleNameBit, cat(one(str("0"))), cat(one(str("1"))))

	// char is the rule `CHAR = %x01-7F`.
	char = newRule(ruleNameChar, cat(one(numRange("x", "01", "7F"))))

	// cr is the rule `CR = %x0D`.
	cr = newRule(ruleNameCr, cat(one(numSeries("x", "0D"))))

	// crlf is the rule `CRLF = CR LF`.
	crlf = newRule(ruleNameCrlf, cat(one(ref(ruleNameCr)), one(ref(ruleNameLf))))

	// ctl is the rule `CTL = %x00-1F / %x7F`.
	ctl = newRule(ruleNameCtl, cat(one(numRange("x", "00", "1F"))), cat(one(numSeries("x", "7F"))))

	// digit is the rule `DIGIT = %x30-39`.
	digit = newRule(ruleNameDigit, cat(one(numRange("x", "30", "39"))))

	// dquote is the rule `DQUOTE = %x22`.
	dquote = newRule(ruleNameDquote, cat(one(numSeries("x", "22"))))

	// hexdig is the rule `HEXDIG = DIGIT / "A" / "B" / "C" / "D" / "E" / "F"`.
	hexdig = newRule(ruleNameHexdig, cat(one(ref(ruleNameDigit))), cat(one(str("A"))), cat(one(str("B"))), cat(one(str("C"))), cat(one(str("D"))), cat(one(str("E"))), cat(one(str("F"))))

	// htab is the rule `HTAB = %x09`.
	htab = newRule(ruleNameHtab, cat(one(numSeries("x", "09"))))

	// lf is the rule `LF = %x0A`.
	lf = newRule(ruleNameLf, cat(one(numSeries("x", "0A"))))

	// lwsp is the rule `LWSP = *(WSP / CRLF WSP)`.
	lwsp = newRule(ruleNameLwsp, cat(rep(0, inf, grp(cat(one(ref(ruleNameWsp))), cat(one(ref(ruleNameCrlf)), one(ref(ruleNameWsp)))))))

	// octet is the rule `OCTET = %x00-FF`.
	octet = newRule(ruleNameOctet, cat(one(numRange("x", "00", "FF"))))

	// sp is the rule `SP = %x20`.
	sp = newRule(ruleNameSp, cat(one(numSeries("x", "20"))))

	// vchar is the rule `VCHAR = %x21-7E`.
	vchar = newRule(ruleNameVchar, cat(one(numRange("x", "21", "7E"))))

	// wsp is the rule `WSP = SP / HTAB`.
	wsp = newRule(ruleNameWsp, cat(one(ref(ruleNameSp))), cat(one(ref(ruleNameHtab))))
)

// coreRules indexes the core rules by lowercase rule name.
var coreRules = map[string]*Rule{
	"alpha":  alpha,
	"bit":    bit,
	"char":   char,
	"cr":     cr,
	"crlf":   crlf,
	"ctl":    ctl,
	"digit":  digit,
	"dquote": dquote,
	"hexdig": hexdig,
	"htab":   htab,
	"lf":     lf,
	"lwsp":   lwsp,
	"octet":  octet,
	"sp":     sp,
	"vchar":  vchar,
	"wsp":    wsp,
}

// The ABNF grammar of ABNF (RFC 5234 Section 4, with the RFC 7405
// char-val extension and applied errata).
var (
	// abnfRulelist is the rule `rulelist = 1*(rule / (*WSP c-nl))`.
	abnfRulelist = newRule(ruleNameRulelist, cat(
		rep(
			1,
			inf,
			grp(
				cat(one(ref(ruleNameRule))),
				cat(one(grp(cat(rep(0, inf, ref(ruleNameWsp)), one(ref(ruleNameCNl)))))),
			),
		),
	))

	// abnfRule is the rule `rule = rulename defined-as elements c-nl`.
	abnfRule = newRule(ruleNameRule, cat(
		one(ref(ruleNameRulename)),
		one(ref(ruleNameDefinedAs)),
		one(ref(ruleNameElements)),
		one(ref(ruleNameCNl)),
	))

	// abnfRulename is the rule `rulename = ALPHA *(ALPHA / DIGIT / "-")`.
	abnfRulename = newRule(ruleNameRulename, cat(
		one(ref(ruleNameAlpha)),
		rep(0, inf, grp(cat(one(ref(ruleNameAlpha))), cat(one(ref(ruleNameDigit))), cat(one(str("-"))))),
	))

	// abnfDefinedAs is the rule `defined-as = *c-wsp ("=" / "=/") *c-wsp`.
	abnfDefinedAs = newRule(ruleNameDefinedAs, cat(
		rep(0, inf, ref(ruleNameCWsp)),
		one(grp(cat(one(str("="))), cat(one(str("=/"))))),
		rep(0, inf, ref(ruleNameCWsp)),
	))

	// abnfElements is the rule `elements = alternation *WSP`.
	abnfElements = newRule(ruleNameElements, cat(one(ref(ruleNameAlternation)), rep(0, inf, ref(ruleNameWsp))))

	// abnfCWsp is the rule `c-wsp = WSP / (c-nl WSP)`.
	abnfCWsp = newRule(ruleNameCWsp, cat(one(ref(ruleNameWsp))), cat(one(grp(cat(one(ref(ruleNameCNl)), one(ref(ruleNameWsp)))))))

	// abnfCNl is the rule `c-nl = comment / CRLF`.
	abnfCNl = newRule(ruleNameCNl, cat(one(ref(ruleNameComment))), cat(one(ref(ruleNameCrlf))))

	// abnfComment is the rule `comment = ";" *(WSP / VCHAR) CRLF`.
	abnfComment = newRule(ruleNameComment, cat(
		one(str(";")),
		rep(0, inf, grp(cat(one(ref(ruleNameWsp))), cat(one(ref(ruleNameVchar))))),
		one(ref(ruleNameCrlf)),
	))

	// abnfAlternation is the rule `alternation = concatenation *(*c-wsp "/" *c-wsp concatenation)`.
	abnfAlternation = newRule(ruleNameAlternation, cat(
		one(ref(ruleNameConcatenation)),
		rep(
			0,
			inf,
			grp(
				cat(
					rep(0, inf, ref(ruleNameCWsp)),
					one(str("/")),
					rep(0, inf, ref(ruleNameCWsp)),
					one(ref(ruleNameConcatenation)),
				),
			),
		),
	))

	// abnfConcatenation is the rule `concatenation = repetition *(1*c-wsp repetition)`.
	abnfConcatenation = newRule(ruleNameConcatenation, cat(
		one(ref(ruleNameRepetition)),
		rep(0, inf, grp(cat(rep(1, inf, ref(ruleNameCWsp)), one(ref(ruleNameRepetition))))),
	))

	// abnfRepetition is the rule `repetition = [repeat] element`.
	abnfRepetition = newRule(ruleNameRepetition, cat(one(opt(cat(one(ref(ruleNameRepeat))))), one(ref(ruleNameElement))))

	// abnfRepeat is the rule `repeat = 1*DIGIT / (*DIGIT "*" *DIGIT)`.
	abnfRepeat = newRule(ruleNameRepeat, cat(rep(1, inf, ref(ruleNameDigit))), cat(one(grp(cat(rep(0, inf, ref(ruleNameDigit)), one(str("*")), rep(0, inf, ref(ruleNameDigit)))))))

	// abnfElement is the rule `element = rulename / group / option / char-val / num-val / prose-val`.
	abnfElement = newRule(ruleNameElement, cat(one(ref(ruleNameRulename))), cat(one(ref(ruleNameGroup))), cat(one(ref(ruleNameOption))), cat(one(ref(ruleNameCharVal))), cat(one(ref(ruleNameNumVal))), cat(one(ref(ruleNameProseVal))))

	// abnfGroup is the rule `group = "(" *c-wsp alternation *c-wsp ")"`.
	abnfGroup = newRule(ruleNameGroup, cat(
		one(str("(")),
		rep(0, inf, ref(ruleNameCWsp)),
		one(ref(ruleNameAlternation)),
		rep(0, inf, ref(ruleNameCWsp)),
		one(str(")")),
	))

	// abnfOption is the rule `option = "[" *c-wsp alternation *c-wsp "]"`.
	abnfOption = newRule(ruleNameOption, cat(
		one(str("[")),
		rep(0, inf, ref(ruleNameCWsp)),
		one(ref(ruleNameAlternation)),
		rep(0, inf, ref(ruleNameCWsp)),
		one(str("]")),
	))

	// abnfCharVal is the rule `char-val = case-insensitive-string / case-sensitive-string`.
	abnfCharVal = newRule(ruleNameCharVal, cat(one(ref(ruleNameCaseInsensitiveString))), cat(one(ref(ruleNameCaseSensitiveString))))

	// abnfCaseInsensitiveString is the rule `case-insensitive-string = ["%i"] quoted-string`.
	abnfCaseInsensitiveString = newRule(ruleNameCaseInsensitiveString, cat(one(opt(cat(one(str("%i"))))), one(ref(ruleNameQuotedString))))

	// abnfCaseSensitiveString is the rule `case-sensitive-string = "%s" quoted-string`.
	abnfCaseSensitiveString = newRule(ruleNameCaseSensitiveString, cat(one(str("%s")), one(ref(ruleNameQuotedString))))

	// abnfQuotedString is the rule `quoted-string = DQUOTE *(%x20-21 / %x23-7E) DQUOTE`.
	abnfQuotedString = newRule(ruleNameQuotedString, cat(
		one(ref(ruleNameDquote)),
		rep(0, inf, grp(cat(one(numRange("x", "20", "21"))), cat(one(numRange("x", "23", "7E"))))),
		one(ref(ruleNameDquote)),
	))

	// abnfNumVal is the rule `num-val = "%" (bin-val / dec-val / hex-val)`.
	abnfNumVal = newRule(ruleNameNumVal, cat(
		one(str("%")),
		one(grp(cat(one(ref(ruleNameBinVal))), cat(one(ref(ruleNameDecVal))), cat(one(ref(ruleNameHexVal))))),
	))

	// abnfBinVal is the rule `bin-val = "b" 1*BIT [1*("." 1*BIT) / ("-" 1*BIT)]`.
	abnfBinVal = newRule(ruleNameBinVal, cat(
		one(str("b")),
		rep(1, inf, ref(ruleNameBit)),
		one(
			opt(
				cat(rep(1, inf, grp(cat(one(str(".")), rep(1, inf, ref(ruleNameBit)))))),
				cat(one(grp(cat(one(str("-")), rep(1, inf, ref(ruleNameBit)))))),
			),
		),
	))

	// abnfDecVal is the rule `dec-val = "d" 1*DIGIT [1*("." 1*DIGIT) / ("-" 1*DIGIT)]`.
	abnfDecVal = newRule(ruleNameDecVal, cat(
		one(str("d")),
		rep(1, inf, ref(ruleNameDigit)),
		one(
			opt(
				cat(rep(1, inf, grp(cat(one(str(".")), rep(1, inf, ref(ruleNameDigit)))))),
				cat(one(grp(cat(one(str("-")), rep(1, inf, ref(ruleNameDigit)))))),
			),
		),
	))

	// abnfHexVal is the rule `hex-val = "x" 1*HEXDIG [1*("." 1*HEXDIG) / ("-" 1*HEXDIG)]`.
	abnfHexVal = newRule(ruleNameHexVal, cat(
		one(str("x")),
		rep(1, inf, ref(ruleNameHexdig)),
		one(
			opt(
				cat(rep(1, inf, grp(cat(one(str(".")), rep(1, inf, ref(ruleNameHexdig)))))),
				cat(one(grp(cat(one(str("-")), rep(1, inf, ref(ruleNameHexdig)))))),
			),
		),
	))

	// abnfProseVal is the rule `prose-val = "<" *(%x20-3D / %x3F-7E) ">"`.
	abnfProseVal = newRule(ruleNameProseVal, cat(
		one(str("<")),
		rep(0, inf, grp(cat(one(numRange("x", "20", "3D"))), cat(one(numRange("x", "3F", "7E"))))),
		one(str(">")),
	))
)

// abnfGrammar is the grammar used to parse ABNF grammar definitions.
var abnfGrammar = &Grammar{Rulemap: map[string]*Rule{
	ruleNameRulelist:              abnfRulelist,
	ruleNameRule:                  abnfRule,
	ruleNameRulename:              abnfRulename,
	ruleNameDefinedAs:             abnfDefinedAs,
	ruleNameElements:              abnfElements,
	ruleNameCWsp:                  abnfCWsp,
	ruleNameCNl:                   abnfCNl,
	ruleNameComment:               abnfComment,
	ruleNameAlternation:           abnfAlternation,
	ruleNameConcatenation:         abnfConcatenation,
	ruleNameRepetition:            abnfRepetition,
	ruleNameRepeat:                abnfRepeat,
	ruleNameElement:               abnfElement,
	ruleNameGroup:                 abnfGroup,
	ruleNameOption:                abnfOption,
	ruleNameCharVal:               abnfCharVal,
	ruleNameCaseInsensitiveString: abnfCaseInsensitiveString,
	ruleNameCaseSensitiveString:   abnfCaseSensitiveString,
	ruleNameQuotedString:          abnfQuotedString,
	ruleNameNumVal:                abnfNumVal,
	ruleNameBinVal:                abnfBinVal,
	ruleNameDecVal:                abnfDecVal,
	ruleNameHexVal:                abnfHexVal,
	ruleNameProseVal:              abnfProseVal,
}}
