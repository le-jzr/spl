SPL
===

*NOTE: This document is unfinished.*

SPL is a simple format for representing structured data.
It is heavily inspired by Ron Rivest's S-Expressions (http://people.csail.mit.edu/rivest/sexp.html).

An SPL file is a sequence of SPL objects.
An SPL object is one of the following:
 * STRING: a text string (an ordered sequence of unicode code points)
 * INTEGER: a signed integer of arbitrary magnitude
 * BLOB: a byte array (an ordered sequence of unsigned 8-bit integers)
 * LIST: a list of zero or more SPL objects

There are no references. SPL objects are always finite, non-recursive, tree-structured, and have a well-defined tree depth.
Also, the STRING type must not contain the NUL character, i.e. the code point with value 0.

A possible representation as a C type
-------------------------------------

	typedef some_integer_representation SPL_int;

	typedef enum {
		SPL_STRING, SPL_INTEGER, SPL_BLOB, SPL_LIST
	} SPL_kind;

	typedef struct SPL_object { 
		SPL_kind tag;
		union {
			char *string;
			SPL_int integer;
			uint8_t *blob;
			struct SPL_object *list;
		};
	} SPL_object;


Printable text representation
-----------------------------

When an SPL object needs to be stored in a human-readable text (e.g. as a configuration file), the following format is to be used:

STRING:
	The string is encoded by escaping it and enclosing it in double quotes.
	Before enclosing, non-printable characters, double quotes and backslashes ale all transformed into escape sequences.
	Escape sequences supported are: \", \\, \t, \n, \xHH, \uHHHH, \UHHHHHHHH. All in their traditional meaning, except for \n,
	which always represents the application's preferred representation of a "new line" mark.
	\x escape sequence is interpreted as a byte in UTF-8 format. A sequence of adjacent bytes is decoded as a single UTF-8 string.
	It is permitted to encode a single unicode code point as a sequence of its constituent UTF-8 bytes.

INTEGER:
	The integer value is encoded as a conventional decimal number string, with no spaces or other separators, and no enclosing characters.
	Example: The number -12458 would be represented using the character sequence '-', '1', '2', '4', '5', '8'.

BLOB:
	The byte array is represented by the character '#', followed by the length of the array in decimal notation,
	followed by the character ":" followed by the bytes of the array, each encoded as two hexadecimal digits in lowercase.
	Example: The array { 0, 1, 26, 87, 128, 13 } would be represented as '#6:00011a57800d'.
	
LIST:
	The object list is represented as a sequence of representations of its constituent objects separated by whitespace,
	the entire sequence enclosed in a pair of regular parentheses.
	Example: A list of string "hello", string "world", integer 1337, an empty list, and a blob {0,1,1,2,3,5,8,13},
	         is represented as '("hello" "world" 1337 () #8:000101020305080d)'.



Binary stream representation
----------------------------

For streaming over a network, a different representation is defined.
This representation is designed for efficient transfer of data from one application to another, with as little transformation as possible.
If you do not care for efficiency, you can just use the text representation.
However, do not use the binary representation anywhere, where you expect a human to read it.

First off, the representation has two parameters.
The endianity (little-endian/big-endian), and the string encoding (UTF-8, UTF-16, whatever else).
In absence of any other requirements or negotiation on the application level, the recommended defaults are little-endian and UTF-8.
Rationale: UTF-8 is space-efficient for strings in latin-derived writing systems.
Little-endian is for consistency with long integer encoding, where little-endian is simpler to work with.

Before any other data, the stream starts with a LIST of up to 128 key STRINGs, as the first SPL object.
The set of key strings may be empty. It depends on the application protocol or a statistical analysis of data,
and serves to assign a single-byte identifiers to most frequent strings.


Definition of Raw integer:
	Raw integer is a variable-length encoding of an unsigned integer.
	It is a sequence of bytes. The most significant bit of each byte
	is the end marker. If it is set to 0, the next byte is a
	continuation of raw integer. If it is 1, the byte is the last byte of
	the integer. The resulting value is computed by appending the 7 bits
	from each byte in little-endian order.

Each SPL object is encoded by first writing its total length in bytes as
a Raw integer.

STRING:
	When sending a string, if the string is one of the predefined key strings, it can be sent as a single byte
	indicating the (zero-based) index of the string in the key string list.
	If not one of the key strings, the string is sent as a byte 0x80, followed by three bytes encoding the length
	of the string's encoding in bytes (in server's preferred endianity), followed by the bytes of the encoding
	(no escaping or delimiting characters).

INTEGER:
	
*TODO*


Extensions, clarifications, etc.
--------------------------------

Q: What character encoding is the printable text representation?
A: Unimportant, but you are bound to make someone angry if you don't use UTF-8.

Q: How do you encode date/time/datetime information?
A: String (https://tools.ietf.org/html/rfc3339), or integer (https://en.wikipedia.org/wiki/Unix_time) are both good choices. Make your pick and make it stick.

Q: How do you encode anything else?
A: Any way you want. If you think something deserves a guideline here, let me know.
