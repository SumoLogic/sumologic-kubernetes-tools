# Logs generator

Logs generator is tool for generating logs (text lines) using patterns,
which can specify changing parts (words, digits)

## General options

Logs generation can be control by multiple options:

- `duration` - defines how much time in seconds  logs should be genereated.
- `total-logs` - defines how many logs in total should be generated.
- `throughput` - maximum throughput (bytes per second).
- `logs-throughput` - maximum number of logs generated per second.

## Patterns

Generator uses patterns to generate logs.
Pattern is basically a log line with placeholders used to differentiate logs.

Currently supported placeholders:

- `{w}` is going to be replaced with random word from wordlist
- `{d}` is going to be replaced with random digit
- `{c}` is going to be replaced with logs counter

Example patterns:

```text
log number is {c}: {w} is random word and {d} is random digit
todays digits are: {d} {d} {d} {d} {d}
this log is always the same
```

### Patterns Generation

Patterns can be randomly generated using given options:

- `random-patterns` - number of patterns to be generated
- `min` - minimum length of pattern (in words)
- `max` - maximum length of pattern (in words)
- `known-words` - ratio of known words
- `random-words` - ratio of random words (`{w}` placeholder)
- `random-digits` - ratio of random digits (`{d}` placeholder)

Ratio is calculated from all of those values using following formula:
`<value>/(<known-words> + <random-words> + <random-digits>)`

### Providing patterns

There are three ways of providing patterns into generator:

- `pattern-file` - including patterns from file (one per line)
- `pattern` - including patterns from string (separated with `$`) eg, `--pattern='{d}${w}'`,
  mind `'` to avoid `$` evaluation
- `random-patterns` - generate patterns randomly. See [patterns generation](#patterns-generation)
