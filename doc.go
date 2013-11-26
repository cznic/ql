// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//MAYBE set operations
//MAYBE IN ( SELECT ... )
//MAYBE +=, -=, ...

//TODO verify there's a graceful failure for a 2G+ blob on a 32 bit machine.

// Package ql is a pure Go embedded (S)QL database.
//
// QL is a SQL-like language. It is less complex and less powerful than SQL
// (whichever specification SQL is considered to be).
//
// Release notes
//
// This is a first public release. More features will probably be added in
// later releases while considering user feedback, if any.
//
// Notation
//
// The syntax is specified using Extended Backus-Naur Form (EBNF)
//
// 	Production  = production_name "=" [ Expression ] "." .
// 	Expression  = Alternative { "|" Alternative } .
// 	Alternative = Term { Term } .
// 	Term        = production_name | token [ "…" token ] | Group | Option | Repetition .
// 	Group       = "(" Expression ")" .
// 	Option      = "[" Expression "]" .
// 	Repetition  = "{" Expression "}" .
// 	Productions are expressions constructed from terms and the following operators, in increasing precedence
//
// 	|   alternation
// 	()  grouping
// 	[]  option (0 or 1 times)
// 	{}  repetition (0 to n times)
//
// Lower-case production names are used to identify lexical tokens.
// Non-terminals are in CamelCase. Lexical tokens are enclosed in double quotes
// "" or back quotes ``.
//
// The form a … b represents the set of characters from a through b as
// alternatives. The horizontal ellipsis … is also used elsewhere in the spec
// to informally denote various enumerations or code snippets that are not
// further specified.
//
// QL source code representation
//
// QL source code is Unicode text encoded in UTF-8. The text is not
// canonicalized, so a single accented code point is distinct from the same
// character constructed from combining an accent and a letter; those are
// treated as two code points.  For simplicity, this document will use the
// unqualified term character to refer to a Unicode code point in the source
// text.
//
// Each code point is distinct; for instance, upper and lower case letters are
// different characters.
//
// Implementation restriction: For compatibility with other tools, the parser
// may disallow the NUL character (U+0000) in the statement.
//
// Implementation restriction: A byte order mark is disallowed anywhere in QL
// statements.
//
// Characters
//
// The following terms are used to denote specific character classes
//
//  newline        = . // the Unicode code point U+000A
//  unicode_char   = . // an arbitrary Unicode code point except newline
//  ascii_letter   = "a" … "z" | "A" … "Z" .
//
// Letters and digits
//
// The underscore character _ (U+005F) is considered a letter.
//
//  letter        = ascii_letter | "_" .
//  decimal_digit = "0" … "9" .
//  octal_digit   = "0" … "7" .
//  hex_digit     = "0" … "9" | "A" … "F" | "a" … "f" .
//
// Lexical elements
//
// Lexical elements are comments, tokens, identifiers, keywords, operators and
// delimiters, integer, floating-point, imaginary, rune and string literals and
// QL parameters.
//
// Comments
//
// There are three forms of comments
//
// Line comments start with the character sequence // or -- and stop at the end
// of the line. A line comment acts like a space.
//
// General comments start with the character sequence /* and continue through
// the character sequence */. A general comment acts like a space.
//
// Comments do not nest.
//
// Tokens
//
// Tokens form the vocabulary of QL. There are four classes: identifiers,
// keywords, operators and delimiters, and literals. White space, formed from
// spaces (U+0020), horizontal tabs (U+0009), carriage returns (U+000D), and
// newlines (U+000A), is ignored except as it separates tokens that would
// otherwise combine into a single token.
//
// Semicolons
//
// The formal grammar uses semicolons ";" as separators of QL statements. A
// single QL statement or the last QL statement in a list of statements can
// have an optional semicolon terminator. (Actually a separator from the
// following empty statement.)
//
// Identifiers
//
// Identifiers name entities such as tables or record set columns. An
// identifier is a sequence of one or more letters and digits. The first
// character in an identifier must be a letter.
//
//  identifier = letter { letter | decimal_digit } .
//
// For example
//
// 	price
// 	_tmp42
// 	Sales
//
// No identifiers are predeclared, however note that no keyword can be used as
// an identifier.
//
// Keywords
//
// The following keywords are reserved and may not be used as identifiers.
//
//	ADD    BETWEEN  BY          CREATE    duration  FROM    int16  NOT     TABLE     uint16  VALUES
//	ALTER  bigint   byte        DELETE    false     GROUP   int32  NULL    time      uint32  WHERE
//	AND    bigrat   COLUMN      DESC      float     IN      int64  ORDER   true      uint64
//	AS     blob     complex128  DISTINCT  float32   INSERT  int8   SELECT  TRUNCATE  uint8
//	ASC    bool     complex64   DROP      float64   int     INTO   string  uint      UPDATE
//
// Keywords are not case sensitive.
//
// Operators and Delimiters
//
// The following character sequences represent operators, delimiters, and other
// special tokens
//
// 	+    &    &&    ==    !=    (    )
// 	-    |    ||    <     <=    [    ]
// 	*    ^          >     >=    ,    ;
// 	/    <<         =           .
// 	%    >>         !
// 	     &^
//
// Operators consisting of more than one character are referred to by names in
// the rest of the documentation
//
//  andand = "&&" .
//  andnot = "&^" .
//  lsh    = "<<" .
//  le     = "<=" .
//  eq     = "==" .
//  ge     = ">=" .
//  neq    = "!=" .
//  oror   = "||" .
//  rsh    = ">>" .
//
// Integer literals
//
// An integer literal is a sequence of digits representing an integer constant.
// An optional prefix sets a non-decimal base: 0 for octal, 0x or 0X for
// hexadecimal.  In hexadecimal literals, letters a-f and A-F represent values
// 10 through 15.
//
//  int_lit     = decimal_lit | octal_lit | hex_lit .
//  decimal_lit = ( "1" … "9" ) { decimal_digit } .
//  octal_lit   = "0" { octal_digit } .
//  hex_lit     = "0" ( "x" | "X" ) hex_digit { hex_digit } .
//
// For example
//
// 	42
// 	0600
// 	0xBadFace
// 	1701411834604692
//
// Floating-point literals
//
// A floating-point literal is a decimal representation of a floating-point
// constant. It has an integer part, a decimal point, a fractional part, and an
// exponent part. The integer and fractional part comprise decimal digits; the
// exponent part is an e or E followed by an optionally signed decimal
// exponent.  One of the integer part or the fractional part may be elided; one
// of the decimal point or the exponent may be elided.
//
//  float_lit = decimals "." [ decimals ] [ exponent ] |
//              decimals exponent |
//              "." decimals [ exponent ] .
//  decimals  = decimal_digit { decimal_digit } .
//  exponent  = ( "e" | "E" ) [ "+" | "-" ] decimals .
//
// For example
//
// 	0.
// 	72.40
// 	072.40  // == 72.40
// 	2.71828
// 	1.e+0
// 	6.67428e-11
// 	1E6
// 	.25
// 	.12345E+5
//
// Imaginary literals
//
// An imaginary literal is a decimal representation of the imaginary part of a
// complex constant. It consists of a floating-point literal or decimal integer
// followed by the lower-case letter i.
//
//  imaginary_lit = (decimals | float_lit) "i" .
//
// For example
//
// 	0i
// 	011i  // == 11i
// 	0.i
// 	2.71828i
// 	1.e+0i
// 	6.67428e-11i
// 	1E6i
// 	.25i
// 	.12345E+5i
//
// Rune literals
//
// A rune literal represents a rune constant, an integer value identifying a
// Unicode code point. A rune literal is expressed as one or more characters
// enclosed in single quotes. Within the quotes, any character may appear
// except single quote and newline. A single quoted character represents the
// Unicode value of the character itself, while multi-character sequences
// beginning with a backslash encode values in various formats.
//
// The simplest form represents the single character within the quotes; since
// QL statements are Unicode characters encoded in UTF-8, multiple
// UTF-8-encoded bytes may represent a single integer value. For instance, the
// literal 'a' holds a single byte representing a literal a, Unicode U+0061,
// value 0x61, while 'ä' holds two bytes (0xc3 0xa4) representing a literal
// a-dieresis, U+00E4, value 0xe4.
//
// Several backslash escapes allow arbitrary values to be encoded as ASCII
// text.  There are four ways to represent the integer value as a numeric
// constant: \x followed by exactly two hexadecimal digits; \u followed by
// exactly four hexadecimal digits; \U followed by exactly eight hexadecimal
// digits, and a plain backslash \ followed by exactly three octal digits. In
// each case the value of the literal is the value represented by the digits in
// the corresponding base.
//
// Although these representations all result in an integer, they have different
// valid ranges. Octal escapes must represent a value between 0 and 255
// inclusive.  Hexadecimal escapes satisfy this condition by construction. The
// escapes \u and \U represent Unicode code points so within them some values
// are illegal, in particular those above 0x10FFFF and surrogate halves.
//
// After a backslash, certain single-character escapes represent special
// values
//
// 	\a   U+0007 alert or bell
// 	\b   U+0008 backspace
// 	\f   U+000C form feed
// 	\n   U+000A line feed or newline
// 	\r   U+000D carriage return
// 	\t   U+0009 horizontal tab
// 	\v   U+000b vertical tab
// 	\\   U+005c backslash
// 	\'   U+0027 single quote  (valid escape only within rune literals)
// 	\"   U+0022 double quote  (valid escape only within string literals)
//
// All other sequences starting with a backslash are illegal inside rune
// literals.
//
//  rune_lit         = "'" ( unicode_value | byte_value ) "'" .
//  unicode_value    = unicode_char | little_u_value | big_u_value | escaped_char .
//  byte_value       = octal_byte_value | hex_byte_value .
//  octal_byte_value = `\` octal_digit octal_digit octal_digit .
//  hex_byte_value   = `\` "x" hex_digit hex_digit .
//  little_u_value   = `\` "u" hex_digit hex_digit hex_digit hex_digit .
//  big_u_value      = `\` "U" hex_digit hex_digit hex_digit hex_digit
//                             hex_digit hex_digit hex_digit hex_digit .
//  escaped_char     = `\` ( "a" | "b" | "f" | "n" | "r" | "t" | "v" | `\` | "'" | `"` ) .
//
// For example
//
// 	'a'
// 	'ä'
// 	'本'
// 	'\t'
// 	'\000'
// 	'\007'
// 	'\377'
// 	'\x07'
// 	'\xff'
// 	'\u12e4'
// 	'\U00101234'
// 	'aa'         // illegal: too many characters
// 	'\xa'        // illegal: too few hexadecimal digits
// 	'\0'         // illegal: too few octal digits
// 	'\uDFFF'     // illegal: surrogate half
// 	'\U00110000' // illegal: invalid Unicode code point
//
// String literals
//
// A string literal represents a string constant obtained from concatenating a
// sequence of characters. There are two forms: raw string literals and
// interpreted string literals.
//
// Raw string literals are character sequences between back quotes ``. Within
// the quotes, any character is legal except back quote. The value of a raw
// string literal is the string composed of the uninterpreted (implicitly
// UTF-8-encoded) characters between the quotes; in particular, backslashes
// have no special meaning and the string may contain newlines. Carriage
// returns inside raw string literals are discarded from the raw string value.
//
// Interpreted string literals are character sequences between double quotes
// "".  The text between the quotes, which may not contain newlines, forms the
// value of the literal, with backslash escapes interpreted as they are in rune
// literals (except that \' is illegal and \" is legal), with the same
// restrictions. The three-digit octal (\nnn) and two-digit hexadecimal (\xnn)
// escapes represent individual bytes of the resulting string; all other
// escapes represent the (possibly multi-byte) UTF-8 encoding of individual
// characters. Thus inside a string literal \377 and \xFF represent a single
// byte of value 0xFF=255, while ÿ, \u00FF, \U000000FF and \xc3\xbf represent
// the two bytes 0xc3 0xbf of the UTF-8 encoding of character U+00FF.
//
//  string_lit             = raw_string_lit | interpreted_string_lit .
//  raw_string_lit         = "`" { unicode_char | newline } "`" .
//  interpreted_string_lit = `"` { unicode_value | byte_value } `"` .
//
// For example
//
// 	`abc`  // same as "abc"
// 	`\n
// 	\n`    // same as "\\n\n\\n"
// 	"\n"
// 	""
// 	"Hello, world!\n"
// 	"日本語"
// 	"\u65e5本\U00008a9e"
// 	"\xff\u00FF"
// 	"\uD800"       // illegal: surrogate half
// 	"\U00110000"   // illegal: invalid Unicode code point
//
// These examples all represent the same string
//
// 	"日本語"                                 // UTF-8 input text
// 	`日本語`                                 // UTF-8 input text as a raw literal
// 	"\u65e5\u672c\u8a9e"                    // the explicit Unicode code points
// 	"\U000065e5\U0000672c\U00008a9e"        // the explicit Unicode code points
// 	"\xe6\x97\xa5\xe6\x9c\xac\xe8\xaa\x9e"  // the explicit UTF-8 bytes
//
// If the statement source represents a character as two code points, such as a
// combining form involving an accent and a letter, the result will be an error
// if placed in a rune literal (it is not a single code point), and will appear
// as two code points if placed in a string literal.
//
// QL parameters
//
// Literals are assigned their values from the respective text representation
// at "compile" (parse) time. QL parameters provide the same functionality as
// literals, but their value is assigned at execution time from an expression
// list passed to DB.Run or DB.Execute. Using '?' or '$' is completely
// equivalent.
//
//  ql_parameter = ( "?" | "$" ) "0" … "9" { "0" … "9" } .
//
// For example
//
// 	SELECT DepartmentID
// 	FROM department
// 	WHERE DepartmentID == ?1
// 	ORDER BY DepartmentName;
//
// 	SELECT employee.LastName
// 	FROM department, employee
// 	WHERE department.DepartmentID == $1 && employee.LastName > $2
// 	ORDER BY DepartmentID;
//
// Constants
//
// Keywords 'false' and 'true' (not case sensitive) represent the two possible
// constant values of type bool (also not case sensitive).
//
// Keyword 'NULL' (not case sensitive) represents an untyped constant which is
// assignable to any type. NULL is distinct from any other value of any type.
//
// Types
//
// A type determines the set of values and operations specific to values of
// that type. A type is specified by a type name.
//
//  Type = "bigint"      // http://golang.org/pkg/math/big/#Int
//       | "bigrat"      // http://golang.org/pkg/math/big/#Rat
//       | "blob"        // []byte
//       | "bool"
//       | "byte"        // alias for uint8
//       | "complex128"
//       | "complex64"
//       | "duration"    // http://golang.org/pkg/time/#Duration
//       | "float"       // alias for float64
//       | "float32"
//       | "float64"
//       | "int"         // alias for int64
//       | "int16"
//       | "int32"
//       | "int64"
//       | "int8"
//       | "rune"        // alias for int32
//       | "string"
//       | "time"        // http://golang.org/pkg/time/#Time
//       | "uint"        // alias for uint64
//       | "uint16"
//       | "uint32"
//       | "uint64"
//       | "uint8" .
//
// Named instances of the boolean, numeric, and string types are keywords. The
// names are not case sensitive.
//
// Note: The blob type is exchanged between the back end and the API as []byte.
// On 32 bit platforms this limits the size which the implementation can handle
// to 2G.
//
// Boolean types
//
// A boolean type represents the set of Boolean truth values denoted by the
// predeclared constants true and false. The predeclared boolean type is bool.
//
// Duration type
//
// A duration type represents the elapsed time between two instants as an int64
// nanosecond count. The representation limits the largest representable
// duration to approximately 290 years.
//
// Numeric types
//
// A numeric type represents sets of integer or floating-point values. The
// predeclared architecture-independent numeric types are
//
// 	uint8       the set of all unsigned  8-bit integers (0 to 255)
// 	uint16      the set of all unsigned 16-bit integers (0 to 65535)
// 	uint32      the set of all unsigned 32-bit integers (0 to 4294967295)
// 	uint64      the set of all unsigned 64-bit integers (0 to 18446744073709551615)
//
// 	int8        the set of all signed  8-bit integers (-128 to 127)
// 	int16       the set of all signed 16-bit integers (-32768 to 32767)
// 	int32       the set of all signed 32-bit integers (-2147483648 to 2147483647)
// 	int64       the set of all signed 64-bit integers (-9223372036854775808 to 9223372036854775807)
// 	duration    the set of all signed 64-bit integers (-9223372036854775808 to 9223372036854775807)
//	bigint      the set of all integers
//
//	bigrat      the set of all rational numbers
//
// 	float32     the set of all IEEE-754 32-bit floating-point numbers
// 	float64     the set of all IEEE-754 64-bit floating-point numbers
//
// 	complex64   the set of all complex numbers with float32 real and imaginary parts
// 	complex128  the set of all complex numbers with float64 real and imaginary parts
//
// 	byte        alias for uint8
// 	float       alias for float64
// 	int         alias for int64
// 	rune        alias for int32
// 	uint        alias for uint64
//
// The value of an n-bit integer is n bits wide and represented using two's
// complement arithmetic.
//
// Conversions are required when different numeric types are mixed in an
// expression or assignment.
//
// String types
//
// A string type represents the set of string values. A string value is a
// (possibly empty) sequence of bytes. The case insensitive keyword for the
// string type is 'string'.
//
// The length of a string (its size in bytes) can be discovered using the
// built-in function len.
//
// Time types
//
// A time type represents an instant in time with nanosecond precision. Each
// time has associated with it a location, consulted when computing the
// presentation form of the time.
//
// Predeclared functions
//
// The following functions are implicitly declared
//
// 	avg complex count date id imag len min max real sum
//
// Expressions
//
// An expression specifies the computation of a value by applying operators and
// functions to operands.
//
// Operands
//
// Operands denote the elementary values in an expression. An operand may be a
// literal, a (possibly qualified) identifier denoting a constant or a function
// or a table/record set column, or a parenthesized expression.
//
//  Operand = Literal | QualifiedIdent | "(" Expression ")" .
//  Literal = "FALSE" | "NULL" | "TRUE"
//  	| float_lit | imaginary_lit | int_lit | rune_lit | string_lit
//  	| ql_parameter .
//
// Qualified identifiers
//
// A qualified identifier is an identifier qualified with a table/record set
// name prefix.
//
//  QualifiedIdent = identifier [ "." identifier ] .
//
// For example
//
// 	invoice.Num	// might denote column 'Num' from table 'invoice'
//
// Primary expressions
//
// Primary expression are the operands for unary and binary expressions.
//
//  PrimaryExpression = Operand
//              | Conversion
//              | PrimaryExpression Index
//              | PrimaryExpression Slice
//              | PrimaryExpression Call .
//
//  Call  = "(" [ ExpressionList ] ")" .
//  Index = "[" Expression "]" .
//  Slice = "[" [ Expression ] ":" [ Expression ] "]" .
//
// For example
//
// 	x
// 	2
// 	(s + ".txt")
// 	f(3.1415, true)
// 	s[i : j + 1]
//
// Index expressions
//
// A primary expression of the form
//
// 	s[x]
//
// denotes the element of a string indexed by x. It's type is byte. The value x
// is called the index.  The following rules apply
//
// - The index x must be of integer type except bigint or duration; it is in
// range if 0 <= x < len(s), otherwise it is out of range.
//
// - A constant index must be non-negative and representable by a value of type
// int.
//
// - A constant index must be in range if the string a is a literal.
//
// - If x is out of range at run time, a run-time error occurs.
//
// - s[x] is the byte at index x and the type of s[x] is byte.
//
// If s is NULL or x is NULL then the result is NULL.
//
// Otherwise s[x] is illegal.
//
// Slices
//
// For a string, the primary expression
//
// 	s[low : high]
//
// constructs a substring. The indices low and high select which elements
// appear in the result. The result has indices starting at 0 and length equal
// to high - low.
//
// For convenience, any of the indices may be omitted. A missing low index
// defaults to zero; a missing high index defaults to the length of the sliced
// operand
//
// 	s[2:]  // same s[2 : len(s)]
// 	s[:3]  // same as s[0 : 3]
// 	s[:]   // same as s[0 : len(s)]
//
// The indices low and high are in range if 0 <= low <= high <= len(a),
// otherwise they are out of range. A constant index must be non-negative and
// representable by a value of type int. If both indices are constant, they
// must satisfy low <= high. If the indices are out of range at run time, a
// run-time error occurs.
//
// Integer values of type bigint or duration cannot be used as indices.
//
// If s is NULL the result is NULL. If low or high is not omitted and is NULL
// then the result is NULL.
//
// Calls
//
// Given an identifier f denoting a predeclared function,
//
// 	f(a1, a2, … an)
//
// calls f with arguments a1, a2, … an. Arguments are evaluated before the
// function is called. The type of the expression is the result type of f.
//
// 	complex(x, y)
// 	len(name)
//
// In a function call, the function value and arguments are evaluated in the
// usual order. After they are evaluated, the parameters of the call are passed
// by value to the function and the called function begins execution. The
// return value of the function is passed by value when the function returns.
//
// Calling an undefined function causes a compile-time error.
//
// Operators
//
// Operators combine operands into expressions.
//
//  Expression = Term { oror Term } .
//
//  ExpressionList = Expression { "," Expression } [ "," ].
//  Factor =  PrimaryFactor  { ( ge | ">" | le | "<" | neq | eq ) PrimaryFactor } [ Predicate ] .
//  PrimaryFactor = PrimaryTerm  { ( "^" | "|" | "-" | "+" ) PrimaryTerm } .
//  PrimaryTerm = UnaryExpr { ( andnot | "&" | lsh | rsh | "%" | "/" | "*" ) UnaryExpr } .
//  Term = Factor { andand Factor } .
//  UnaryExpr = [ "^" | "!" | "-" | "+" ] PrimaryExpression .
//
// Comparisons are discussed elsewhere. For other binary operators, the operand
// types must be identical unless the operation involves shifts or untyped
// constants. For operations involving constants only, see the section on
// constant expressions.
//
// Except for shift operations, if one operand is an untyped constant and the
// other operand is not, the constant is converted to the type of the other
// operand.
//
// The right operand in a shift expression must have unsigned integer type or
// be an untyped constant that can be converted to unsigned integer type. If
// the left operand of a non-constant shift expression is an untyped constant,
// the type of the constant is what it would be if the shift expression were
// replaced by its left operand alone.
//
// Predicates
//
// Predicates are special form expressions having a boolean result type.
// Expressions of the form
//
//	expr IN ( expr1, expr2, expr3, ... )		// case A
//
//	expr NOT IN ( expr1, expr2, expr3, ... )	// case B
//
// are equivalent, including NULL handling, to
//
//	expr == expr1 || expr == expr2 || expr == expr3 || ...	// case A
//
//	expr != expr1 && expr != expr2 && expr != expr3 && ...	// case B
//
// The types of involved expressions must be comparable as defined in
// "Comparison operators".
//
// Expressions of the form
//
//	expr BETWEEN low AND high	// case A
//
//	expr NOT BETWEEN low AND high	// case B
//
// are equivalent, including NULL handling, to
//
//	expr >= low && expr <= high	// case A
//
//	expr < low || expr > high	// case B
//
// The types of involved expressions must be ordered as defined in "Comparison
// operators".
//
//  Predicate = (
//  			[ "NOT" ] (
//  			  "IN" "(" ExpressionList ")"
//  			| "BETWEEN" PrimaryFactor "AND" PrimaryFactor
//  			)
//		|
//  			"IS" [ "NOT" ] "NULL"
//  	).
//
// Expressions of the form
//
//	expr IS NULL		// case A
//
//	expr IS NOT NULL	// case B
//
// yeild a boolean value true if expr does not have a specific type (case A) or
// if expr has a specific type (case B). In other cases the result is a boolean
// value false.
//
// Operator precedence
//
// Unary operators have the highest precedence.
//
// There are five precedence levels for binary operators. Multiplication
// operators bind strongest, followed by addition operators, comparison
// operators, && (logical AND), and finally || (logical OR)
//
// 	Precedence    Operator
// 	    5             *  /  %  <<  >>  &  &^
// 	    4             +  -  |  ^
// 	    3             ==  !=  <  <=  >  >=
// 	    2             &&
// 	    1             ||
//
// Binary operators of the same precedence associate from left to right. For
// instance, x / y * z is the same as (x / y) * z.
//
// 	+x
// 	23 + 3*x[i]
// 	x <= f()
// 	^a >> b
// 	f() || g()
// 	x == y+1 && z > 0
//
// Note that the operator precedence is reflected explicitly by the grammar.
//
// Arithmetic operators
//
// Arithmetic operators apply to numeric values and yield a result of the same
// type as the first operand. The four standard arithmetic operators (+, -, *,
// /) apply to integer, rational, floating-point, and complex types; + also
// applies to strings; +,- also applies to times.  All other arithmetic
// operators apply to integers only.
//
// 	+    sum                    integers, rationals, floats, complex values, strings
// 	-    difference             integers, rationals, floats, complex values, times
// 	*    product                integers, rationals, floats, complex values
// 	/    quotient               integers, rationals, floats, complex values
// 	%    remainder              integers
//
// 	&    bitwise AND            integers
// 	|    bitwise OR             integers
// 	^    bitwise XOR            integers
// 	&^   bit clear (AND NOT)    integers
//
// 	<<   left shift             integer << unsigned integer
// 	>>   right shift            integer >> unsigned integer
//
// Strings can be concatenated using the + operator
//
// 	"hi" + string(c) + " and good bye"
//
// String addition creates a new string by concatenating the operands.
//
// A value of type time can be added with a value of type duration.
//
//	now() + duration("1h")	// time after 1 hour from now
//
// Times can subtracted from each other producing a value of type duration.
//
//	now() - t0	// elapsed time since t0
//
// For two integer values x and y, the integer quotient q = x / y and remainder
// r = x % y satisfy the following relationships
//
// 	x = q*y + r  and  |r| < |y|
//
// with x / y truncated towards zero ("truncated division").
//
// 	 x     y     x / y     x % y
// 	 5     3       1         2
// 	-5     3      -1        -2
// 	 5    -3      -1         2
// 	-5    -3       1        -2
//
// As an exception to this rule, if the dividend x is the most negative value
// for the int type of x, the quotient q = x / -1 is equal to x (and r = 0).
//
// 				 x, q
// 	int8                     -128
// 	int16                  -32768
// 	int32             -2147483648
// 	int64    -9223372036854775808
//
// If the divisor is a constant expression, it must not be zero. If the divisor
// is zero at run time, a run-time error occurs. If the dividend is
// non-negative and the divisor is a constant power of 2, the division may be
// replaced by a right shift, and computing the remainder may be replaced by a
// bitwise AND operation
//
// 	 x     x / 4     x % 4     x >> 2     x & 3
// 	 11      2         3         2          3
// 	-11     -2        -3        -3          1
//
// The shift operators shift the left operand by the shift count specified by
// the right operand. They implement arithmetic shifts if the left operand is a
// signed integer and logical shifts if it is an unsigned integer. There is no
// upper limit on the shift count. Shifts behave as if the left operand is
// shifted n times by 1 for a shift count of n. As a result, x << 1 is the same
// as x*2 and x >> 1 is the same as x/2 but truncated towards negative
// infinity.
//
// For integer operands, the unary operators +, -, and ^ are defined as follows
//
// 	+x                          is 0 + x
// 	-x    negation              is 0 - x
// 	^x    bitwise complement    is m ^ x  with m = "all bits set to 1" for unsigned x
// 	                                      and  m = -1 for signed x
//
// For floating-point and complex numbers, +x is the same as x, while -x is the
// negation of x. The result of a floating-point or complex division by zero is
// not specified beyond the IEEE-754 standard; whether a run-time error occurs
// is implementation-specific.
//
// Whenever any operand of any arithmetic operation, unary or binary, is NULL,
// as well as in the case of the string concatenating operation, the result is
// NULL.
//
//	42*NULL		// the result is NULL
//	NULL/x		// the result is NULL
//	"foo"+NULL	// the result is NULL
//
// Integer overflow
//
// For unsigned integer values, the operations +, -, *, and << are computed
// modulo 2n, where n is the bit width of the unsigned integer's type. Loosely
// speaking, these unsigned integer operations discard high bits upon overflow,
// and expressions may rely on ``wrap around''.
//
// For signed integers with a finite bit width, the operations +, -, *, and <<
// may legally overflow and the resulting value exists and is deterministically
// defined by the signed integer representation, the operation, and its
// operands. No exception is raised as a result of overflow. An evaluator may
// not optimize an expression under the assumption that overflow does not
// occur. For instance, it may not assume that x < x + 1 is always true.
//
// Integers of type bigint and rationals do not overflow but their handling is
// limited by the memory resources available to the program.
//
// Comparison operators
//
// Comparison operators compare two operands and yield a boolean value.
//
// 	==    equal
// 	!=    not equal
// 	<     less
// 	<=    less or equal
// 	>     greater
// 	>=    greater or equal
//
// In any comparison, the first operand must be of same type as is the second
// operand, or vice versa.
//
// The equality operators == and != apply to operands that are comparable. The
// ordering operators <, <=, >, and >= apply to operands that are ordered.
// These terms and the result of the comparisons are defined as follows
//
// - Boolean values are comparable. Two boolean values are equal if they are
// either both true or both false.
//
// - Complex values are comparable. Two complex values u and v are equal if
// both real(u) == real(v) and imag(u) == imag(v).
//
// - Integer values are comparable and ordered, in the usual way. Note that
// durations are integers.
//
// - Floating point values are comparable and ordered, as defined by the
// IEEE-754 standard.
//
// - Rational values are comparable and ordered, in the usual way.
//
// - String values are comparable and ordered, lexically byte-wise.
//
// - Time values are comparable and ordered.
//
// Whenever any operand of any comparison operation is NULL, the result is
// NULL.
//
// Note that slices are always of type string.
//
// Logical operators
//
// Logical operators apply to boolean values and yield a boolean result. The
// right operand is evaluated conditionally.
//
// 	&&    conditional AND    p && q  is  "if p then q else false"
// 	||    conditional OR     p || q  is  "if p then true else q"
// 	!     NOT                !p      is  "not p"
//
// The truth tables for logical operations with NULL values
//
// 	+-------+-------+---------+---------+
// 	|   p   |   q   |  p || q |  p && q |
// 	+-------+-------+---------+---------+
// 	| true  | true  | *true   |  true   |
// 	| true  | false | *true   |  false  |
// 	| true  | NULL  | *true   |  NULL   |
// 	| false | true  |  true   | *false  |
// 	| false | false |  false  | *false  |
// 	| false | NULL  |  NULL   | *false  |
// 	| NULL  | true  |  true   |  NULL   |
// 	| NULL  | false |  NULL   |  false  |
// 	| NULL  | NULL  |  NULL   |  NULL   |
// 	+-------+-------+---------+---------+
// 	 * indicates q is not evaluated.
//
// 	+-------+-------+
// 	|   p   |  !p   |
// 	+-------+-------+
// 	| true  | false |
// 	| false | true  |
// 	| NULL  | NULL  |
// 	+-------+-------+
//
// Conversions
//
// Conversions are expressions of the form T(x) where T is a type and x is an
// expression that can be converted to type T.
//
//  Conversion = Type "(" Expression ")" .
//
// A constant value x can be converted to type T in any of these cases:
//
// - x is representable by a value of type T.
//
// - x is a floating-point constant, T is a floating-point type, and x is
// representable by a value of type T after rounding using IEEE 754
// round-to-even rules. The constant T(x) is the rounded value.
//
// - x is an integer constant and T is a string type. The same rule as for
// non-constant x applies in this case.
//
// Converting a constant yields a typed constant as result.
//
// 	float32(2.718281828)     // 2.718281828 of type float32
// 	complex128(1)            // 1.0 + 0.0i of type complex128
// 	float32(0.49999999)      // 0.5 of type float32
// 	string('x')              // "x" of type string
// 	string(0x266c)           // "♬" of type string
// 	"foo" + "bar"            // "foobar"
// 	int(1.2)                 // illegal: 1.2 cannot be represented as an int
// 	string(65.0)             // illegal: 65.0 is not an integer constant
//
// A non-constant value x can be converted to type T in any of these cases:
//
// - x has type T.
//
// - x's type and T are both integer or floating point types.
//
// - x's type and T are both complex types.
//
// - x is an integer, except bigint or duration, and T is a string type.
//
// Specific rules apply to (non-constant) conversions between numeric types or
// to and from a string type. These conversions may change the representation
// of x and incur a run-time cost. All other conversions only change the type
// but not the representation of x.
//
// A conversion of NULL to any type yields NULL.
//
// Conversions between numeric types
//
// For the conversion of non-constant numeric values, the following rules
// apply
//
// 1. When converting between integer types, if the value is a signed integer,
// it is sign extended to implicit infinite precision; otherwise it is zero
// extended.  It is then truncated to fit in the result type's size. For
// example, if v == uint16(0x10F0), then uint32(int8(v)) == 0xFFFFFFF0. The
// conversion always yields a valid value; there is no indication of overflow.
//
// 2. When converting a floating-point number to an integer, the fraction is
// discarded (truncation towards zero).
//
// 3. When converting an integer or floating-point number to a floating-point
// type, or a complex number to another complex type, the result value is
// rounded to the precision specified by the destination type. For instance,
// the value of a variable x of type float32 may be stored using additional
// precision beyond that of an IEEE-754 32-bit number, but float32(x)
// represents the result of rounding x's value to 32-bit precision. Similarly,
// x + 0.1 may use more than 32 bits of precision, but float32(x + 0.1) does
// not.
//
// In all non-constant conversions involving floating-point or complex values,
// if the result type cannot represent the value the conversion succeeds but
// the result value is implementation-dependent.
//
// Conversions to and from a string type
//
// 1. Converting a signed or unsigned integer value to a string type yields a
// string containing the UTF-8 representation of the integer. Values outside
// the range of valid Unicode code points are converted to "\uFFFD".
//
// 	string('a')       // "a"
// 	string(-1)        // "\ufffd" == "\xef\xbf\xbd"
// 	string(0xf8)      // "\u00f8" == "ø" == "\xc3\xb8"
// 	string(0x65e5)    // "\u65e5" == "日" == "\xe6\x97\xa5"
//
// 2. Converting a blob to a string type yields a string whose successive bytes
// are the elements of the blob.
//
//	string(b /* []byte{'h', 'e', 'l', 'l', '\xc3', '\xb8'} */)   // "hellø"
//	string(b /* []byte{} */)                                     // ""
//	string(b /* []byte(nil) */)                                  // ""
//
// 3. Converting a value of a string type to a blob yields a blob whose
// successive elements are the bytes of the string.
//
//	blob("hellø")   // []byte{'h', 'e', 'l', 'l', '\xc3', '\xb8'}
//	blob("")        // []byte{}
//
// 4. Converting a value of a bigint type to a string yields a string
// containing the decimal decimal representation of the integer.
//
//	string(M9)	// "2305843009213693951"
//
// 5. Converting a value of a string type to a bigint yields a bigint value
// containing the integer represented by the string value. A prefix of “0x” or
// “0X” selects base 16; the “0” prefix selects base 8, and a “0b” or “0B”
// prefix selects base 2. Otherwise the value is interpreted in base 10. An
// error occurs if the string value is not in any valid format.
//
//	bigint("2305843009213693951")		// M9
//	bigint("0x1ffffffffffffffffffffff")	// M10 == 2^89-1
//
// 6. Converting a value of a rational type to a string yields a string
// containing the decimal decimal representation of the rational in the form
// "a/b" (even if b == 1).
//
//	string(bigrat(355)/bigrat(113))	// "355/113"
//
// 7. Converting a value of a string type to a bigrat yields a bigrat value
// containing the rational represented by the string value. The string can be
// given as a fraction "a/b" or as a floating-point number optionally followed
// by an exponent. An error occurs if the string value is not in any valid
// format.
//
//	bigrat("1.2e-34")
//	bigrat("355/113")
//
// 8. Converting a value of a duration type to a string returns a string
// representing the duration in the form "72h3m0.5s". Leading zero units are
// omitted. As a special case, durations less than one second format using a
// smaller unit (milli-, micro-, or nanoseconds) to ensure that the leading
// digit is non-zero. The zero duration formats as 0, with no unit.
//
//	string(elapsed)	// "1h", for example
//
// 9. Converting a string value to a duration yields a duration represented by
// the string.  A duration string is a possibly signed sequence of decimal
// numbers, each with optional fraction and a unit suffix, such as "300ms",
// "-1.5h" or "2h45m". Valid time units are "ns", "us" (or "µs"), "ms", "s",
// "m", "h".
//
//	duration("1m")	// http://golang.org/pkg/time/#Minute
//
// 10. Converting a time value to a string returns the time formatted using the
// format string
//
//	"2006-01-02 15:04:05.999999999 -0700 MST"
//
// Order of evaluation
//
// When evaluating the operands of an expression or of function calls,
// operations are evaluated in lexical left-to-right order.
//
// For example, in the evaluation of
//
// 	g(h(), i()+x[j()], c)
//
// the function calls and evaluation of c happen in the order h(), i(), j(), c.
//
// Floating-point operations within a single expression are evaluated according
// to the associativity of the operators. Explicit parentheses affect the
// evaluation by overriding the default associativity. In the expression x + (y
// + z) the addition y + z is performed before adding x.
//
// Statements
//
// Statements control execution.
//
//  Statement =  EmptyStmt | AlterTableStmt | BeginTransactionStmt | CommitStmt
//  	| CreateTableStmt | DeleteFromStmt | DropTableStmt | InsertIntoStmt
//  	| RollbackStmt | SelectStmt | TruncateTableStmt | UpdateStmt .
//
//  StatementList = Statement { ";" Statement } .
//
// Empty statements
//
// The empty statement does nothing.
//
//  EmptyStmt = .
//
// ALTER TABLE
//
// Alter table statements modify existing tables.  With the ADD clause it adds
// a new column to the table. The column must not exist. With the DROP clause
// it removes an existing column from a table. The column must exist.
//
//  AlterTableStmt = "ALTER" "TABLE" TableName ( "ADD" ColumnDef | "DROP" "COLUMN"  ColumnName ) .
//
// For example
//
//	BEGIN TRANSACTION;
// 		ALTER TABLE Stock ADD Qty int;
// 		ALTER TABLE Income DROP COLUMN Taxes;
//	COMMIT;
//
// BEGIN TRANSACTION
//
// Begin transactions statements introduce a new transaction level. Every
// transaction level must be eventually balanced by exactly one of COMMIT or
// ROLLBACK statements. Note that when a transaction is roll-backed because of
// a statement failure then no explicit balancing of the respective BEGIN
// TRANSACTION is statement is required nor permitted.
//
// Failure to properly balance any opened transaction level may cause dead
// locks and/or lose of data updated in the uppermost opened but never properly
// closed transaction level.
//
//  BeginTransactionStmt = "BEGIN" "TRANSACTION" .
//
// For example
//
//	BEGIN TRANSACTION;
//		INSERT INTO foo VALUES (42, 3.14);
//		INSERT INTO foo VALUES (-1, 2.78);
//	COMMIT;
//
// COMMIT
//
// The commit statement closes the innermost transaction nesting level. If
// that's the outermost level then the updates to the DB made by the
// transaction are atomically made persistent.
//
//  CommitStmt = "COMMIT" .
//
// For example
//
//	BEGIN TRANSACTION;
//		INSERT INTO AccountA (Amount) VALUES ($1);
//		INSERT INTO AccountB (Amount) VALUES (-$1);
//	COMMIT;
//
// CREATE TABLE
//
// Create table statements create new tables. A column definition declares the
// column name and type. Table names and column names are case sensitive. The
// table must not exist before.
//
//  CreateTableStmt = "CREATE" "TABLE" TableName "(" ColumnDef { "," ColumnDef } [ "," ] ")" .
//
//  ColumnDef = ColumnName Type .
//  ColumnName = identifier .
//  TableName = identifier .
//
// For example
//
//	BEGIN TRANSACTION;
// 		CREATE TABLE department
// 		(
// 			DepartmentID   int,
// 			DepartmentName string,
// 		);
// 		CREATE TABLE employee
// 		(
// 			LastName	string,
// 			DepartmentID	int,
// 		);
//	COMMIT;
//
//
// DELETE FROM
//
// Delete from statements remove rows from a table, which must exist.
//
//  DeleteFromStmt = "DELETE" "FROM" TableName [ WhereClause ] .
//
// For example
//
//	BEGIN TRANSACTION;
//		DELETE FROM DepartmentID
//		WHERE DepartmentName == "Ponies";
//	COMMIT;
//
// If the WhereClause is not present then all rows are removed and the
// statement is equivalent to the TRUNCATE TABLE statement.
//
// DROP TABLE
//
// Drop table statements remove tables from the DB. The table must exist.
//
//  DropTableStmt = "DROP" "TABLE" TableName .
//
// For example
//
//	BEGIN TRANSACTION;
// 		DROP TABLE Inventory;
//	COMMIT;
//
// INSERT INTO
//
// Insert into statements insert new rows into tables. New rows come from
// literal data, if using the VALUES clause, or are a result of select
// statement. In the later case the select statement is fully evaluated before
// the insertion of any rows is performed, allowing to insert values calculated
// from the same table rows are to be inserted into. If the ColumnNameList part
// is omitted then the number of values inserted in the row must be the same as
// are columns in the table. If the ColumnNameList part is present then the
// number of values per row must be same as the same number of column names.
// All other columns of the record are set to NULL.  The type of the value
// assigned to a column must be the same as is the column's type or the value
// must be NULL.
//
//  InsertIntoStmt = "INSERT" "INTO" TableName [ "(" ColumnNameList ")" ] ( Values | SelectStmt ) .
//
//  ColumnNameList = ColumnName { "," ColumnName } [ "," ] .
//  Values = "VALUES" "(" ExpressionList ")" { "," "(" ExpressionList ")" } [ "," ] .
//
// For example
//
//	BEGIN TRANSACTION;
// 		INSERT INTO department (DepartmentID) VALUES (42);
//
// 		INSERT INTO department (
// 			DepartmentName,
// 			DepartmentID,
// 		)
// 		VALUES (
// 			"R&D",
// 			42,
// 		);
//
// 		INSERT INTO department VALUES
//			(42, "R&D"),
//			(17, "Sales"),
// 		;
//	COMMIT;
//
//	BEGIN TRANSACTION;
//		INSERT INTO department (DepartmentName, DepartmentID)
//		SELECT DepartmentName+"/headquarters", DepartmentID+1000
//		FROM department;
//	COMMIT;
//
// ROLLBACK
//
// The rollback statement closes the innermost transaction nesting level
// discarding any updates to the DB made by it. If that's the outermost level
// then the effects on the DB are as if the transaction never happened.
//
//  RollbackStmt = "ROLLBACK" .
//
// For example
//
//	// First statement list
//	BEGIN TRANSACTION
//		SELECT * INTO tmp FROM foo;
//		INSERT INTO tmp SELECT * from bar;
//		SELECT * from tmp;
//
// The (temporary) record set from the last statement is returned and can be
// processed by the client.
//
//	// Second statement list
//	ROLLBACK;
//
// In this case the rollback is the same as 'DROP TABLE tmp;' but it can be a
// more complex operation.
//
// SELECT FROM
//
// Select from statements produce recordsets. The optional DISTINCT modifier
// ensures all rows in the result recordset are unique. Either all of the
// resulting fields are returned ('*') or only those named in FieldList.
//
// RecordSetList is a list of table names or parenthesized select statements,
// optionally (re)named using the AS clause.
//
// The result can be filtered using a WhereClause and orderd by the OrderBy
// clause.
//
//  SelectStmt = "SELECT" [ "DISTINCT" ] ( "*" | FieldList ) "FROM" RecordSetList
//  	[ WhereClause ] [ GroupByClause ] [ OrderBy ] .
//
//  RecordSet = ( TableName | "(" SelectStmt [ ";" ] ")" ) [ "AS" identifier ] .
//  RecordSetList = RecordSet { "," RecordSet } [ "," ] .
//
// For example
//
// 	SELECT * FROM Stock;
//
// 	SELECT DepartmentID
// 	FROM department
// 	WHERE DepartmentID == 42
// 	ORDER BY DepartmentName;
//
// 	SELECT employee.LastName
// 	FROM department, employee
// 	WHERE department.DepartmentID == employee.DepartmentID
// 	ORDER BY DepartmentID;
//
// If Recordset is a nested, parenthesized SelectStmt then it must be given a
// name using the AS clause if its field are to be accessible in expressions.
//
// 	SELECT a.b, c.d
// 	FROM
// 		x AS a,
// 		(
// 			SELECT * FROM y;
// 		) AS c
//	WHERE a.e > c.e;
//
// Fields naming rules
//
// A field is an named expression. Identifiers, not used as a type in
// conversion or a function name in the Call clause, denote names of (other)
// fields, values of which should be used in the expression.
//
//  Field = Expression [ "AS" identifier ] .
//
// The expression can be named using the AS clause.  If the AS clause is not
// present and the expression consists solely of a field name, then that field
// name is used as the name of the resulting field. Otherwise the field is
// unnamed.
//
// For example
//
//	SELECT 314, 42 as AUQLUE, DepartmentID, DepartmentID+1000, LastName as Name from employee;
//	// Fields are []string{"", "AUQLUE", "DepartmentID", "", "Name"}
//
// The SELECT statement can optionally enumerate the desired/resulting fields
// in a list.
//
//  FieldList = Field { "," Field } [ "," ] .
//
// No two identical field names can appear in the list.
//
//	SELECT DepartmentID, LastName, DepartmentID from employee;
//	// duplicate field name "DepartmentID"
//
//	SELECT DepartmentID, LastName, DepartmentID as ID2 from employee;
//	// works
//
// When more than one record set is used in the FROM clause record set list,
// the result record set field names are rewritten to be qualified using
// the record set names.
//
//	SELECT * FROM employee, department;
//	// Fields are []string{"employee.LastName", "employee.DepartmentID", "department.DepartmentID", "department.DepartmentName"
//
// If a particular record set doesn't have a name, its respective fields became
// unnamed.
//
//	SELECT * FROM employee as e, ( SELECT * FROM department);
//	// Fields are []string{"e.LastName", "e.DepartmentID", "", ""
//
//	SELECT * FROM employee AS e, ( SELECT * FROM department) AS d;
//	// Fields are []string{"e.LastName", "e.DepartmentID", "d.DepartmentID", "d.DepartmentName"
//
// Recordset ordering
//
// Resultins rows of a SELECT statement can be optionally ordered by the ORDER
// BY clause.  Collating proceeds by considering the expressions in the
// expression list left to right until a collating order is determined. Any
// possibly remaining expressions are not evaluated.
//
//  OrderBy = "ORDER" "BY" ExpressionList [ "ASC" | "DESC" ] .
//
// All of the expression values must yield an ordered type or NULL. Ordered
// types are defined in "Comparison operators". Collating of elements having a
// NULL value is different compared to what the comparison operators yield in
// expression evaluation (NULL result instead of a boolean value).
//
// Below, T denotes a non NULL value of any QL type.
//
// 	NULL < T
//
// NULL collates before any non NULL value (is considered smaller than T).
//
//	NULL == NULL
//
// Two NULLs have no collating order (are considered equal).
//
// Recordset filtering
//
// The WHERE clause restricts records considered by some statements, like
// SELECT FROM, DELETE FROM, or UPDATE.
//
//	expression value	consider the record
//	----------------	-------------------
//	true			yes
//	false or NULL		no
//
// It is an error if the expression evaluates to a non null value of non bool
// type.
//
//  WhereClause = "WHERE" Expression .
//
// Recordset grouping
//
// The GROUP BY clause is used to project rows having common values into a
// smaller set of rows.
//
// For example
//
//	SELECT Country, sum(Qty) FROM Sales GROUP BY Country;
//
//	SELECT Country, Product FROM Sales GROUP BY Country, Product;
//
//	SELECT DISTINCT Country, Product FROM Sales;
//
// Using the GROUP BY without any aggregate functions in the selected fields is
// in certain cases equal to using the DISTINCT modifier. The last two examples
// above produce the same resultsets.
//
//  GroupByClause = "GROUP BY" ColumnNameList .
//
// Select statement evaluation order
//
// 1. The FROM clause is evaluated, producing a Cartesian product of its source
// record sets (tables or nested SELECT statements).
//
// 2. If present, the WHERE clause is evaluated.
//
// 3. If present, the GROUP BY clause is evaluated.
//
// 4. The SELECT field expressions are evaluated.
//
// 5. If present, the DISTINCT modifier is evaluated.
//
// 6. If present, the ORDER BY clause is evaluated.
//
// TRUNCATE TABLE
//
// Truncate table statements remove all records from a table. The table must
// exist.
//
//  TruncateTableStmt = "TRUNCATE" "TABLE" TableName .
//
// For example
//
//	BEGIN TRANSACTION
// 		TRUNCATE TABLE department;
//	COMMIT;
//
// UPDATE
//
// Update statements change values of fields in rows of a table.
//
//  UpdateStmt = "UPDATE" TableName AssignmentList [ WhereClause ] .
//
//  AssignmentList = Assignment { "," Assignment } [ "," ] .
//  Assignment = ColumnName "=" Expression .
//
// For example
//
//	BEGIN TRANSACTION
// 		UPDATE department
//			DepartmentName = DepartmentName + " dpt.",
//			DepartmentID = 1000+DepartmentID,
//		WHERE DepartmentID < 1000;
//	COMMIT;
//
// Built-in functions
//
// Built-in functions are predeclared.
//
// Average
//
// The built-in aggregate function avg returns the average of values of an
// expression.  Avg ignores NULL values, but returns NULL if all values of a
// column are NULL or if avg is applied to an empty record set.
//
// 	Call      Argument type    Result
//
// 	avg()     expression       The average of the values of the expression.
//
// The column values must be of a numeric type.
//
//	SELECT salesperson, avg(sales) FROM salesforce GROUP BY salesperson;
//
// Count
//
// The built-in aggregate function count returns how many times an expression
// has a non NULL values or the number of rows in a record set. Note: count()
// returns 0 for an empty record set.
//
// 	Call      Argument type    Result
//
// 	count()   N/A              The number of rows in a record set, int.
// 	count()   expression       The number of cases where the expression value is not NULL, int.
//
// For example
//
//	SELECT count() FROM department; // # of rows
//
//	SELECT count(DepartmentID) FROM department; // # of records with non NULL field DepartmentID
//
//	SELECT count()-count(DepartmentID) FROM department; // # of records with NULL field DepartmentID
//
//	SELECT count(foo+bar*3) AS y FROM t; // # of cases where 'foo+bar*3' is non NULL
//
// Date
//
// Date returns the time corresponding to
//
//	yyyy-mm-dd hh:mm:ss + nsec nanoseconds
//
// in the appropriate zone for that time in the given location.
//
// The month, day, hour, min, sec, and nsec values may be outside their usual
// ranges and will be normalized during the conversion. For example, October 32
// converts to November 1.
//
// A daylight savings time transition skips or repeats times. For example, in
// the United States, March 13, 2011 2:15am never occurred, while November 6,
// 2011 1:15am occurred twice. In such cases, the choice of time zone, and
// therefore the time, is not well-defined. Date returns a time that is correct
// in one of the two zones involved in the transition, but it does not
// guarantee which.
//
// 	Call                                                Result
//
// 	date(year, month, day, hour, min, sec, nsec, loc)   time
//
// The type of all arguments except loc is int. loc must be of type string.  A
// location maps time instants to the zone in use at that time. Typically, the
// location represents the collection of time offsets in use in a geographical
// area, such as "CEST" and "CET" for central Europe.  "local" represents the
// system's local time zone. "UTC" represents Universal Coordinated Time (UTC).
//
// The month specifies a month of the year (January = 1, ...).
//
// If any of the arguments to date is NULL the result is NULL.
//
// Record id
//
// The built-in function id takes no arguments and returns a table-unique
// automatically assigned numeric identifier of type int. Ids of deleted
// records are not reused.
//
// 	Call      Argument type    Result
//
// 	id()      N/A              int
//
// For example
//
//	SELECT id(), LastName
//	FROM employee;
//
// If id() is called for a row which is not a table record then the result
// value is NULL.
//
// For example
//
//	SELECT id(), e.LastName, e.DepartmentID, d.DepartmentID
//	FROM
//		employee AS e,
//		department AS d,
//	WHERE e.DepartmentID == d.DepartmentID;
//	// Will always return NULL in first field.
//
//	SELECT e.ID, e.LastName, e.DepartmentID, d.DepartmentID
//	FROM
//		(SELECT id() AS ID, LastName, DepartmentID FROM employee) AS e,
//		department as d,
//	WHERE e.DepartmentID == d.DepartmentID;
//	// Will work.
//
//
// Length
//
// The built-in function len takes a string argument and returns a result of
// type int.
//
// 	Call      Argument type    Result
//
// 	len(s)    string           string length in bytes
//
// The expression len(s) is constant if s is a string constant.
//
// If the argument to len is NULL the result is NULL.
//
// Maximum
//
// The built-in aggregate function max returns the largest value of an
// expression in a record set.  Max ignores NULL values, but returns NULL if
// all values of a column are NULL or if max is applied to an empty record set.
//
// 	Call      Argument type    Result
//
// 	max()     expression       The largest value of the expression.
//
// The column values must be of an ordered type.
//
// For example
//
//	SELECT department, max(sales) FROM t GROUP BY department;
//
// Minimum
//
// The built-in aggregate function min returns the smallest value of an
// expression in a record set.  Min ignores NULL values, but returns NULL if
// all values of a column are NULL or if min is applied to an empty record set.
//
// 	Call      Argument type    Result
//
// 	min()     expression       The smallest value of the expression.
//
// For example
//
//	SELECT a, min(b) FROM t GROUP BY a;
//
// The column values must be of an ordered type.
//
// Sum
//
// The built-in aggregate function sum returns the sum of values of an
// expression for all rows of a record set. Sum ignores NULL values, but
// returns NULL if all values of a column are NULL or if sum is applied to an
// empty record set.
//
// 	Call      Argument type    Result
//
// 	sum()     expression       The sum of the values of the expression.
//
// The column values must be of a numeric type.
//
//	SELECT salesperson, sum(sales) FROM salesforce GROUP BY salesperson;
//
// Manipulating complex numbers
//
// Three functions assemble and disassemble complex numbers. The built-in
// function complex constructs a complex value from a floating-point real and
// imaginary part, while real and imag extract the real and imaginary parts of
// a complex value.
//
// 	complex(realPart, imaginaryPart floatT) complexT
// 	real(complexT) floatT
// 	imag(complexT) floatT
//
// The type of the arguments and return value correspond. For complex, the two
// arguments must be of the same floating-point type and the return type is the
// complex type with the corresponding floating-point constituents: complex64
// for float32, complex128 for float64. The real and imag functions together
// form the inverse, so for a complex value z, z == complex(real(z), imag(z)).
//
// If the operands of these functions are all constants, the return value is a
// constant.
//
// 	complex(2, -2)             	// complex128
// 	complex(1.0, -1.4)         	// complex128
// 	float32(math.Cos(math.Pi/2))  	// float32
// 	complex(5, float32(-x))         // complex64
// 	imag(b)                   	// float64
// 	real(complex(5, float32(-x)))   // float32
//
// If any argument to any of complex, real, imag functions is NULL the result
// is NULL.
//
// Size guarantees
//
// For the numeric types, the following sizes are guaranteed
//
// 	type                                            size in bytes
//
// 	byte, uint8, int8                                1
// 	uint16, int16                                    2
// 	uint32, int32, float32                           4
// 	uint, uint64, int, int64, float64, complex64     8
// 	complex128                                      16
//
// License
//
// Portions of this specification page are modifications based on work[2]
// created and shared by Google[3] and used according to terms described in the
// Creative Commons 3.0 Attribution License[4].
//
// This specification is licensed under the Creative Commons Attribution 3.0
// License, and code is licensed under a BSD license[5].
//
// References
//
// Links from the above documentation
//
// 	[1]: http://golang.org/ref/spec#Notation
//	[2]: http://golang.org/ref/spec
//	[3]: http://code.google.com/policies.html
//	[4]: http://creativecommons.org/licenses/by/3.0/
//	[5]: http://golang.org/LICENSE
//
package ql
