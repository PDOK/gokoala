/*
 * ------------------
 * Note: This file is based on https://github.com/ldproxy/xtraplatform-spatial/blob/482b607f6709389fcd43ebea7dd0434389b8011b/
 * xtraplatform-cql/src/main/antlr/de/ii/xtraplatform/cql/infra/CqlLexer.g4
 *
 * Keep the following license header intact:
 * ------------------
 *
 * Copyright interactive instruments GmbH
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 */
lexer grammar CqlLexer;

/*
#=============================================================================#
# Enable case-insensitive grammars
#=============================================================================#
*/

fragment A : [aA];
fragment B : [bB];
fragment C : [cC];
fragment D : [dD];
fragment E : [eE];
fragment F : [fF];
fragment G : [gG];
fragment H : [hH];
fragment I : [iI];
fragment J : [jJ];
fragment K : [kK];
fragment L : [lL];
fragment M : [mM];
fragment N : [nN];
fragment O : [oO];
fragment P : [pP];
fragment Q : [qQ];
fragment R : [rR];
fragment S : [sS];
fragment T : [tT];
fragment U : [uU];
fragment V : [vV];
fragment W : [wW];
fragment X : [xX];
fragment Y : [yY];
fragment Z : [zZ];


/*
#=============================================================================#
# Definition of COMPARISON operators
#=============================================================================#
*/

ComparisonOperator : EQ | NEQ | LT | GT | LTEQ | GTEQ;

LT : '<';

EQ : '=';

GT : '>';

NEQ : LT GT;

GTEQ : GT EQ;

LTEQ : LT EQ;



/*
#=============================================================================#
# Definition of BOOLEAN literals
#=============================================================================#
*/

BooleanLiteral : T R U E | F A L S E;

/*
#=============================================================================#
# Definition of LOGICAL operators
#=============================================================================#
*/

AND : A N D;

OR : O R;

NOT : N O T;

/*
#=============================================================================#
# Definition of COMPARISON operators
#=============================================================================#
*/

LIKE : L I K E;

BETWEEN : B E T W E E N;

IS : I S;

NULL: N U L L;

/*
#=============================================================================#
# Definition of SPATIAL operators
#=============================================================================#
*/

SpatialFunction : S UNDERSCORE I N T E R S E C T S
                | S UNDERSCORE E Q U A L S
                | S UNDERSCORE D I S J O I N T
                | S UNDERSCORE T O U C H E S
                | S UNDERSCORE W I T H I N
                | S UNDERSCORE O V E R L A P S
                | S UNDERSCORE C R O S S E S
                | S UNDERSCORE C O N T A I N S;

/*
#=============================================================================#
# Definition of TEMPORAL operators
#=============================================================================#
*/

TemporalFunction : T UNDERSCORE A F T E R
                 | T UNDERSCORE B E F O R E
                 | T UNDERSCORE C O N T A I N S
                 | T UNDERSCORE D I S J O I N T
                 | T UNDERSCORE D U R I N G
                 | T UNDERSCORE E Q U A L S
                 | T UNDERSCORE F I N I S H E D B Y
                 | T UNDERSCORE F I N I S H E S
                 | T UNDERSCORE I N T E R S E C T S
                 | T UNDERSCORE M E E T S
                 | T UNDERSCORE M E T B Y
                 | T UNDERSCORE O V E R L A P P E D B Y
                 | T UNDERSCORE O V E R L A P S
                 | T UNDERSCORE S T A R T E D B Y
                 | T UNDERSCORE S T A R T S;

/*
#=============================================================================#
# Definition of ARRAY operators
#=============================================================================#
*/
ArrayFunction : A UNDERSCORE E Q U A L S
              | A UNDERSCORE C O N T A I N S
              | A UNDERSCORE C O N T A I N E D B Y
              | A UNDERSCORE O V E R L A P S;

/*
#=============================================================================#
# Definition of IN operator
#=============================================================================#
*/

IN: I N;

/*
#=============================================================================#
# Definition of geometry types
#=============================================================================#
*/

POINT: P O I N T (Whitespace)* (Z)?;

LINESTRING: L I N E S T R I N G (Whitespace)* (Z)?;

POLYGON: P O L Y G O N (Whitespace)* (Z)?;

MULTIPOINT: M U L T I P O I N T (Whitespace)* (Z)?;

MULTILINESTRING: M U L T I L I N E S T R I N G (Whitespace)* (Z)?;

MULTIPOLYGON: M U L T I P O L Y G O N (Whitespace)* (Z)?;

GEOMETRYCOLLECTION: G E O M E T R Y C O L L E C T I O N (Whitespace)* (Z)?;

BBOX: B B O X;




CharacterStringLiteralStart : QUOTE -> more, mode(STR);

CASEI: C A S E I;

ACCENTI: A C C E N T I;

LOWER: L O W E R;

UPPER: U P P E R;

NumericLiteral : UnsignedNumericLiteral | SignedNumericLiteral;

DIGIT : [0-9];

DOLLAR : '$';

UNDERSCORE : '_';

DOUBLEQUOTE : '"';

QUOTE : '\'';

LEFTPAREN : '(';

RIGHTPAREN : ')';

LEFTSQUAREBRACKET : '[';

RIGHTSQUAREBRACKET : ']';

ASTERISK : '*';

PLUS : '+';

COMMA : ',';

CARET : '^';

MINUS : '-';

PERIOD : '.';

SOLIDUS : '/';

COLON : ':';

PERCENT : '%';

DIV : D I V;

/*
# character & digit productions copied from:
# https://www.w3.org/TR/REC-xml/#charsets
*/

ALPHA : '\u0007'..'\u0008'     //bell, bs
      | '\u0021'..'\u0026'     //!, ", #, $, %, &
      | '\u0028'..'\u002F'     //(, ), *, +, comma, -, ., /
      | '\u003A'..'\u0084'     // --+
      | '\u0086'..'\u009F'     //   |
      | '\u00A1'..'\u167F'     //   |
      | '\u1681'..'\u1FFF'     //   |
      | '\u200B'..'\u2027'     //   +-> :,;,<,=,>,?,@,A-Z,[,\,],^,_,`,a-z,...
      | '\u202A'..'\u202E'     //   |
      | '\u2030'..'\u205E'     //   |
      | '\u2060'..'\u2FFF'     //   |
      | '\u3001'..'\uD7FF'     // --+
      | '\uE000'..'\uFFFD';    // See note 8.
      // '\u10000'..'\u10FFFF' are not supported in ANTLR lexer sets;

/*
# Note 8: Private Use, CJK Compatibility Ideographs, Alphabetic Presentation
#         Forms, Arabic Presentation Forms-A, Combining Half Marks, CJK
#         Compatibility Forms, Small Form Variants, Arabic Presentation Forms-B,
#         Specials, Halfwidth and Fullwidth Forms, Specials
*/

IdentifierStart : '\u003A'              // colon
                | '\u005F'              // underscore
                | '\u0041'..'\u005A'    // A-Z
                | '\u0061'..'\u007A'    // a-z
                | '\u00C0'..'\u00D6'    // À-Ö Latin-1 Supplement Letters
                | '\u00D8'..'\u00F6'    // Ø-ö Latin-1 Supplement Letters
                | '\u00F8'..'\u02FF'    // ø-ÿ Latin-1 Supplement Letters
                | '\u0370'..'\u037D'    // Ͱ-ͽ Greek and Coptic (without ';')
                | '\u037F'..'\u1FFE'    // See note 1.
                | '\u200C'..'\u200D'    // zero width non-joiner and joiner
                | '\u2070'..'\u218F'    // See note 2.
                | '\u2C00'..'\u2FEF'    // See note 3.
                | '\u3001'..'\uD7FF'    // See note 4.
                | '\uF900'..'\uFDCF'    // See note 5.
                | '\uFDF0'..'\uFFFD';   // See note 6.
                // | '\u10000'..'\uEFFFF'; are not supported in ANTLR lexer sets

IdentifierPart : IdentifierStart
               | DIGIT                  // 0-9
               | '\u0300'..'\u036F'     // combining and diacritical marks
               | '\u203F'..'\u2040';    // ‿ and ⁀

/*
# See: https://unicode-table.com/en/blocks/
# Note 1: Greek, Coptic, Cyrillic, Cyrillic Supplement, Armenian, Hebrew,
#         Arabic, Syriac, Arabic Supplement, Thaana, NKo, Samaritan, Mandaic,
#         Syriac Supplement, Arabic Extended-A, Devanagari, Bengali, Gurmukhi,
#         Gujarati, Oriya, Tamil, Telugu, Kannada, Malayalam, Sinhala, Thai,
#         Lao, Tibetan, Myanmar, Georgian, Hangul Jamo, Ethiopic, Ethiopic
#         Supplement, Cherokee, Unified Canadian Aboriginal Syllabics, Ogham,
#         Runic, Tagalog, Hanunoo, Buhid, Tagbanwa, Khmer, Mongolian, Unified
#         Canadian Aboriginal Syllabics Extended, Limbu, Tai Le, New Tai Lue,
#         Khmer Symbols, Buginese, Tai Tham, Combining Diacritical Marks
#         Extended, Balinese, Sundanese, Batak, Lepcha, Ol Chiki, Cyrillic
#         Extended C, Georgian Extended, Sundanese Supplement, Vedic
#         Extensions, Phonetic Extensions, Phonetic Extensions Supplement,
#         Combining Diacritical Marks Supplement, Latin Extended Additional,
#         Greek Extended
#
# Note 2: Superscripts and Subscripts, Currency Symbols, Combining Diacritical
#         Marks for Symbols, Letterlike Symbols, Number Forms (e.g. Roman
#         numbers)
#
# Note 3: Glagolitic, Latin Extended-C, Coptic, Georgian Supplement, Tifinagh,
#         Ethiopic Extended, Cyrillic Extended-A, Supplemental Punctuation,
#         CJK Radicals Supplement, Kangxi Radicals
#
# Note 4: CJK Symbols and Punctuation Hiragana, Katakana, Bopomofo, Hangul
#         Compatibility Jamo, Kanbun, Bopomofo Extended, CJK Strokes, Katakana
#         Phonetic Extensions, Enclosed CJK Letters and Months, CJK
#         Compatibility, CJK Unified Ideographs Extension A, Yijing Hexagram
#         Symbols, CJK Unified Ideographs, Yi Syllables, Yi Radicals, Lisu,
#         Vai, Cyrillic Extended-B, Bamum, Modifier Tone Letters, Latin
#         Extended-D, Syloti Nagri, Common Indic Number Forms, Phags-pa,
#         Saurashtra, Devanagari Extended, Kayah Li, Rejang, Hangul Jamo
#         Extended-A, Javanese, Myanmar Extended-B, Cham, Myanmar Extended-A,
#         Tai Viet, Meetei Mayek Extensions, Ethiopic Extended-A, Latin
#         Extended-E, Cherokee Supplement, Meetei Mayek, Hangul Syllables,
#         Hangul Jamo Extended-B
#
# Note 5: CJK Compatibility Ideographs, Alphabetic Presentation Forms,
#         Arabic Presentation Forms-A
#
# Note 6: Arabic Presentation Forms-A, Variation Selectors, Vertical Forms,
#         Combining Half Marks, CJK Compatibility Forms, Small Form Variants,
#         Arabic Presentation Forms-B, Halfwidth and Fullwidth Forms, Specials
*/

/*
#=============================================================================#
# Definition of NUMERIC literals
#=============================================================================#
*/

UnsignedNumericLiteral : ExactNumericLiteral | ApproximateNumericLiteral;

SignedNumericLiteral : (Sign)? ExactNumericLiteral | ApproximateNumericLiteral;

ExactNumericLiteral : UnsignedInteger  (PERIOD (UnsignedInteger)? )?
                        |  PERIOD UnsignedInteger;

ApproximateNumericLiteral : Mantissa 'E' Exponent;

Mantissa : ExactNumericLiteral;

Exponent : SignedInteger;

SignedInteger : (Sign)? UnsignedInteger;

UnsignedInteger : (DIGIT)+;

Sign : PLUS | MINUS;

/*
#=============================================================================#
# Definition of TEMPORAL literals
#=============================================================================#
*/


NOW : N O W; // NOW is a CQL2 extension

DATE : D A T E;

TIMESTAMP : T I M E S T A M P;

INTERVAL : I N T E R V A L;

DateString : QUOTE FullDate QUOTE;

TimestampString : QUOTE FullDate 'T' UtcTime QUOTE;

DotDotString : QUOTE '..' QUOTE;

Instant : FullDate | FullDate 'T' UtcTime;

FullDate : DateYear '-' DateMonth '-' DateDay;

DateYear : DIGIT DIGIT DIGIT DIGIT;

DateMonth : '0' [1-9] | '1' [0-2];

DateDay : '0' [1-9] | [1-2] DIGIT | '3' [0-1];

UtcTime : TimeHour ':' TimeMinute ':' TimeSecond Z;

TimeHour : [0-1] DIGIT | '2' [0-3];

TimeMinute : [0-5] DIGIT;

TimeSecond : [0-5] DIGIT (PERIOD (DIGIT)+)?;

/*
#=============================================================================#
# Definition of identifiers (property or function names)
#=============================================================================#
*/

Identifier : IdentifierBare | DOUBLEQUOTE IdentifierBare DOUBLEQUOTE;

// CHANGE: moved PERIOD to propertyName
// CHANGE: support a reduced set of characters as identifiers
IdentifierBare : IdentifierStart (IdentifierPart)*;

/*
#=============================================================================#
# ANTLR ignore whitespace
#=============================================================================#
*/

fragment Whitespace :
             '\u0009'  // Character tabulation
           | '\u000A'  // Line feed
           | '\u000B'  // Line tabulation
           | '\u000C'  // Form feed
           | '\u000D'  // Carriage return
           | '\u0020'  // Space
           | '\u0085'  // Next line
           | '\u00A0'  // No-break space
           | '\u1680'  // Ogham space mark
           | '\u2000'  // En quad
           | '\u2001'  // Em quad
           | '\u2002'  // En space
           | '\u2003'  // Em space
           | '\u2004'  // Three-per-em space
           | '\u2005'  // Four-per-em space
           | '\u2006'  // Six-per-em space
           | '\u2007'  // Figure space
           | '\u2008'  // Punctuation space
           | '\u2009'  // Thin space
           | '\u200A'  // Hair space
           | '\u2028'  // Line separator
           | '\u2029'  // Paragraph separator
           | '\u202F'  // Narrow no-break space
           | '\u205F'  // Medium mathematical space
           | '\u3000'; // Ideographic space

WS : Whitespace+ -> skip; // skip spaces, tabs, newlines, etc.

/*
#=============================================================================#
# ANTLR mode for CharacterStringLiteral with whitespaces
#=============================================================================#
*/

mode STR;

CharacterStringLiteral: '\'' -> mode(DEFAULT_MODE);

QuotedQuote: '\'\'' -> more;

Character : ~['] -> more;