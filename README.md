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

There are no references.
SPL objects are always finite, non-recursive, tree-structured, and have a well-defined tree depth.
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

 * STRING  
	The string is encoded by escaping it and enclosing it in double quotes.
	Before enclosing, non-printable characters, double quotes and backslashes ale all transformed into escape sequences.
	Escape sequences supported are: `\"`, `\\`, `\t`, `\n`, `\xHH`, `\uHHHH`, `\UHHHHHHHH`. All in their traditional meaning, except for `\n`,
	which always represents the application's preferred representation of a "new line" mark.
	`\xHH` escape sequence is interpreted as a byte in UTF-8 format. A sequence of adjacent bytes is decoded as a single UTF-8 string.
	It is permitted to encode a single unicode code point as a sequence of its constituent UTF-8 bytes.

 * INTEGER  
	The integer value is encoded as a conventional decimal number string, with no spaces or other separators, and no enclosing characters.
	Example: The number -12458 would be represented using the character sequence `-`, `1`, `2`, `4`, `5`, `8`.

 * BLOB  
	The byte array is represented by the character `#`, followed by the length of the array in decimal notation,
	followed by the character `:` followed by the bytes of the array, each encoded as two hexadecimal digits in lowercase.
	Example: The array { 0, 1, 26, 87, 128, 13 } would be represented as `#6:00011a57800d`.
	
 * LIST  
	The object list is represented as a sequence of representations of its constituent objects separated by whitespace,
	the entire sequence enclosed in a pair of regular parentheses.
	Example: A list of string "hello", string "world", integer 1337, an empty list, and a blob {0,1,1,2,3,5,8,13},
	         is represented as `("hello" "world" 1337 () #8:000101020305080d)`.



Binary stream representation
----------------------------

For streaming over a network, a different representation is defined.
This representation is designed for efficient transfer of data from one application to another, with as little transformation as possible.
If you do not care for efficiency, you can just use the text representation.
However, do not use the binary representation anywhere, where you expect a human to read it.

Before any other data, the stream starts with a LIST of up to 112 key STRINGs, as the first SPL object.
The set of key strings may be empty. It depends on the application protocol or a statistical analysis of data,
and serves to assign a single-byte identifiers to most frequent strings.

Control bytes:  
 * `0x80`--`0xEF` Key STRINGs.
 * `0xF0` reserved
 * `0xF1` reserved
 * `0xF2` reserved
 * `0xF3` reserved
 * `0xF4` reserved
 * `0xF5` reserved
 * `0xF6` reserved
 * `0xF7` reserved
 * `0xF8` reserved
 * `0xF9` reserved
 * `0xFA` Start of a LIST.
 * `0xFB` End of a LIST.
 * `0xFC` Start of a STRING.
 * `0xFD` Start of a BLOB.
 * `0xFE` Start of a positive INTEGER.
 * `0xFF` Start of a negative INTEGER.

7-bit integer encoding (INT7):  
	INT7 a variable-length encoding of an unsigned integer.
	It is a sequence of bytes. The most significant bit of each byte
	is set to zero, i.e. each byte contains 7 bits of the integer.
	Bytes are in little-endian order, i.e. the first byte contains
	the 7 least significant bits of a number.
	There must be no trailing zero bytes.

Each SPL object is optionally prefixed with the length of the entire object
in INT7 encoding, not including the length of the INT7 length itself, but
including any control bytes.
For objects of type BLOB or INTEGER, this length-prefix is mandatory.
After the length, one control byte identifies the type of the object,
followed by object data, as described bellow.

 * STRING:  
	A byte in range `0x80`--`0xEF`, or a byte of value `0xFC`.
	In the former case, the byte's value minus 128 identifies one of the key strings.
	No further data is included. In the latter case, string data encoded as UTF-8 follow
	the control byte, terminated by a byte of value 0.

 * INTEGER:
	Either `0xFE` (nonnegative) or `0xFF` (negative), followed by little-endian sequence
	of bytes, with no trailing zero bytes. Each byte contains full 8 bits of the absolute
	value of the number. Zero is nonnegative, i.e. `01FE`, not `01FF`, nor `02FE00`.

 * BLOB:
	`0xFD` followed by bytes of the blob. Example for {1, 2, 3}: `04FD010203`. Example for {}: `01FD`.

 * LIST:
	`0xFA`, followed by encodings of the contained objects, followed by `0xFB`.


Extensions, clarifications, etc.
--------------------------------

Q: What character encoding is the printable text representation?
A: Unimportant, but you are bound to make someone angry if you don't use UTF-8.

Q: How do you encode date/time/datetime information?
A: String (https://tools.ietf.org/html/rfc3339), or integer (https://en.wikipedia.org/wiki/Unix_time) are both good choices. Make your pick and make it stick.

Q: How do you encode anything else?
A: Any way you want. If you think something deserves a guideline here, let me know.
