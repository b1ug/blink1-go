# Color values and single-state commands

The SDK returns values that its existing control methods already accept. A
parsed color goes directly to `PlayColor` or `NewLightState`; a parsed command
returns `LightState` directly for `PlayState`. Callers do not need a separate
DTO, a Hex round trip, or manual R/G/B extraction.

```go
state, err := blink1.ParseStateQuery("fade to blue over 2s on led 1")
if err != nil {
    return err
}
return controller.PlayState(state)
```

This fragment assumes an opened controller and a function returning an error.
When opening a device for a command, parse first so invalid input is rejected
before hardware access. Open the controller once when handling multiple commands.

`PlayState` applies the same range/precision checks to typed Go states before
writing, so callers do not need to repeat validation. Missing colors (including
typed nil), invalid LED selectors, and unrepresentable fades return an error.
`PlayColor` also rejects missing colors. Errors from typed values have an empty
source range because no input text was parsed. Low-level raw device commands
and hardware Pattern operations remain separate APIs.

For preflight validation before opening hardware or exporting typed/decoded
state data, call `state.Validate()`. It performs the same checks without I/O.
No separate parser import or duplicate LED/time/color checks are needed.
An ordinary `PlayState(state)` call already does this automatically.

## Colors

`ParseColor` accepts exactly one complete color expression:

| Form | Example | Meaning |
| --- | --- | --- |
| Opaque CSS name | `tomato` | One of 148 standard names. |
| Hex | `#abc`, `#12abef` | Three or six hexadecimal digits with `#`. |
| RGB | `rgb(1, 2, 3)` | Three integer channels from 0 to 255. |
| HSV / HSB | `hsv(210,75,90)` | Hue in degrees, saturation/value in percent. |

Keywords and names are ASCII case-insensitive. HSV and HSB share the same
conversion: H is in `[0,360]`, S/V in `[0,100]`; original decimal components
allow at most six fractional digits. Signs, exponent notation, alpha, clipping,
and out-of-range channels are rejected. `hsb` remains a supported spelling of
the existing color model.

```go
light, err := blink1.ParseColor("hsv(210,75,90)")
if err != nil {
    return err
}
return controller.PlayColor(light)
```

For typed Go values, `colorkit.RGB8` already implements `color.Color`:

```go
light, err := (colorkit.HSV{H: 210, S: 75, V: 90}).RGB8()
if err != nil {
    return err
}
return controller.PlayColor(light)
```

The HSV conversion can fail, so checking its error is intentional. It is the
only conversion needed at the call site. `PlayColor` and `PlayState` respect
the controller's Gamma setting; `PlayRGB` is a raw-channel operation that
bypasses Gamma. Unpacking a color into `PlayRGB` is not an equivalent shortcut.

The existing `HexToColor` helper also accepts unprefixed three/six-digit Hex.
`ParseColor` requires `#` to distinguish a Hex literal from a color name.
Neither color parser searches inside a sentence. `on` and `off` are action
words, not color aliases; use `white` or `black`. Local `@name`, random colors,
and configuration lookup belong to the application.

## Command form

```text
set [led N] [to] COLOR [on led N] [now]
off [on] [led N] [now]
fade [led N] [to] COLOR (over|in) DURATION [on led N]
```

Only one LED selector is allowed. Omission means `LEDAll` (numeric 0); N may
be 0, 1, 2, all, both, first, or second. Physical top/bottom orientation is
not inferred. Set and Off are immediate; Off sets black and does not mean
cancel a host Job.

Fade requires an explicit positive transition duration. Units are `ms`, `s`,
or `m`, with at most six decimal places. The result must be an exact multiple
of 10ms, no greater than 655350ms. Values are not silently rounded or clipped.
`over` and `in` specify transition time, not delayed start. Sentence-final
punctuation and general English prose are outside this finite grammar.

These are complete commands:

```text
set red
set led 1 to #abc
off on led 2
fade to blue over 2s on led 1
```

Negation, unknown words, duplicate fields, holds, sequences and higher-level
effects are rejected. For example, `blink red 3 times` cannot fit one
`LightState`; use nb1's Action interface. `set red for 3s` is a hold request,
not a three-second Fade. The SDK never discards these words and executes a
different command.

## Field form

```text
led=1 color=hsv(210,75,90) fade=20ms
```

The three keys `led`, `color`, and `fade` are all required, in any order,
exactly once. No extra keys are accepted. `fade=0ms` means immediate; positive
values follow the same protocol limits as Fade commands. Function colors can
contain spaces around their components. The field form is another spelling of
one state, not a second player or a way to express hold/repeat.

## Input and diagnostics

Commands are bounded to 4096 bytes of valid UTF-8 on one line. A `//` tail
comment starts at the beginning of input or after whitespace. `#abc` is a
color, not a comment. Comments are only accepted by the state parser, not
inside a complete color expression. Keywords are normalized locally; the
original text remains available for error locations.

Errors return no executable partial result: `ParseColor` returns a nil color,
and `ParseStateQuery` returns a zero `LightState` whose Color is nil. Inspect
diagnostics through the SDK without another parser import:

```go
var detail *blink1.ParseError
if errors.As(err, &detail) {
    fmt.Printf("%s: %s at %d:%d; %s\n",
        detail.Code, detail.Field, detail.Line, detail.Column, detail.Hint)
}
```

`Start`/`End` are a half-open byte range in the original input. `Line` and
`Column` are one-based rune positions. `Code`, `Field`, and the range are for
machines; Message/Hint are English descriptions. Do not match whole error text.

| Code | Meaning |
| --- | --- |
| `invalid_input` | Invalid UTF-8, line shape, size, or compact state data. |
| `missing_field` | A required action, color, selector or duration is absent. |
| `invalid_color` | A color literal or component has invalid syntax. |
| `invalid_duration` | A duration has invalid syntax or cannot use exact 10ms units. |
| `value_out_of_range` | A numeric channel, selector or duration is outside its range. |
| `conflicting_values` | A field is repeated or incompatible with another field. |
| `unsupported_intent` | The request cannot be represented as one supported state. |
| `unexpected_input` | Remaining words or fields are not part of the grammar. |

## Compact data and behavior changes

The established `LightState` text codec `#RRGGBBLnTms` and SDK semicolon
`StateSequence` data format remain available. A data codec can describe values
the hardware cannot exactly play. `ParseStateQuery` accepts an explicit compact
state only after checking its executable transition range and precision. For
example, the data codec can read `T15`; the command parser rejects it rather
than pretending it is a 10ms fade.

The command parser no longer scans loosely worded sentences for three unrelated
fields. Use the documented command or field form. `ParseTitle` and
`ParseRepeatTimes` are separate existing helpers; this single-state change does
not redefine their metadata or looping behavior.

Named lookup is now the same CSS set as colorkit. Four historical SDK names and
their Go constants change: green `#008000`, lime `#00ff00`, plum `#dda0dd`, and
violet `#ee82ee`. For fixed colors, store Hex/RGB rather than a reverse-lookup
name. The historical Go constants ColorApricot, ColorBronze, ColorMint,
ColorPeach and ColorScarlet retain their values; these five names are not CSS
bare color literals.

`RainbowColors` is an artistic seven-color palette with explicit RGB anchors:
`#ff0000`, `#ffa500`, `#ffff00`, `#00ff00`, `#0000ff`, `#4b0082`, `#8f00ff`.
It does not inherit CSS green/violet changes. A circular effect closes its own
cycle; the palette itself has seven entries.

For native Go, use the root SDK entry points shown above. The pure-Go
`textparse` subpackage is available for applications such as WASM compilers
that need parsing without importing the native HID backend. SDK adapters
perform the boundary conversion once; ordinary callers do not need to use
the subpackage or manually construct an intermediate state.
