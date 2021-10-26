#[macro_use]
extern crate clap_v3;

use clap_v3::{App, Arg};
use rand::Rng;
use std::fs::File;
use std::fs::OpenOptions;
use std::io::{BufRead, BufReader, Write};
use std::ops::Sub;
use std::process;
use std::thread;
use std::time::{Duration, Instant};
use std::vec::Vec;

fn main() {
    let matches = App::new("fluent-test")
                            .version("v0.1")

                            .help_heading("General options")
                            .arg(Arg::with_name("path")
                                .short('p')
                                .long("path")
                                .value_name("path")
                                .help("Path to the ouput file (/dev/stdout by default)")
                                .required(false)
                                .takes_value(true))
                            .arg(Arg::with_name("logs-throughput")
                               .short('n')
                               .long("logs-throughput")
                               .value_name("logs-throughput")
                               .help("Maximum number of logs generated per second")
                               .required(false)
                               .takes_value(true))
                            .arg(Arg::with_name("throughput")
                               .short('b')
                               .long("throughput")
                               .value_name("throughput")
                               .help("Maximum throughput (bytes per second)")
                               .required(false)
                               .takes_value(true))
                            .arg(Arg::with_name("total-logs")
                                .short('T')
                                .long("total-logs")
                                .value_name("total-logs")
                                .help("Total number of generated logs")
                                .required(false)
                                .takes_value(true))
                            .arg(Arg::with_name("wordlist")
                                .short('w')
                                .long("wordlist")
                                .value_name("wordlist")
                                .help("Wordlist to use")
                                .required(false)
                                .takes_value(true))
                            .arg(Arg::with_name("verbose")
                                .short('S')
                                .long("verbose")
                                .value_name("verbose")
                                .help("Print additional logs")
                                .required(false)
                                .takes_value(true))
                            .arg(Arg::with_name("duration")
                                .short('i')
                                .long("duration")
                                .value_name("duration")
                                .help("Logs generation duration")
                                .required(false)
                                .takes_value(true))
                            .arg(Arg::with_name("time-resolution")
                                .short('I')
                                .long("time-resolution")
                                .value_name("time-resolution")
                                .help(concat!(
                                    "Time resolution (in seconds) defines how often ",
                                    "throughput should be verified and statistics printed")
                                )
                                .required(false)
                                .takes_value(true))

                            .help_heading("Pattern generation")
                            .arg(Arg::with_name("pattern")
                               .short('x')
                               .long("pattern")
                               .value_name("pattern")
                               .help("Pattern to use. You can use special words {w}, {d} and {c} to include random word, digit and log counter")
                               .required(false)
                               .takes_value(true))
                            .arg(Arg::with_name("pattern-file")
                                .short('L')
                                .long("pattern-file")
                                .value_name("pattern-file")
                                .help("Pattern file to use. Every line is considered as separate pattern")
                                .required(false)
                                .takes_value(true))
                            .arg(Arg::with_name("random-patterns")
                               .short('z')
                               .long("random-patterns")
                               .value_name("random-patterns")
                               .help("Amount of random patterns to generate")
                               .required(false)
                               .takes_value(true))
                            .arg(Arg::with_name("min")
                               .short('m')
                               .long("min")
                               .value_name("min")
                               .help("Minimum random pattern length (words)")
                               .required(false)
                               .takes_value(true))
                            .arg(Arg::with_name("max")
                               .short('M')
                               .long("max")
                               .value_name("max")
                               .help("Maximum random pattern length (words)")
                               .required(false)
                               .takes_value(true))
                            .arg(Arg::with_name("random_words")
                               .short('W')
                               .long("random_words")
                               .value_name("random_words")
                               .help("Define ratio of random words to be used in the pattern")
                               .required(false)
                               .takes_value(true))
                            .arg(Arg::with_name("random_digits")
                               .short('D')
                               .long("random_digits")
                               .value_name("random_digits")
                               .help("Define ratio of random digits to be used in the pattern")
                               .required(false)
                               .takes_value(true))
                            .arg(Arg::with_name("known_words")
                               .short('K')
                               .long("known_words")
                               .value_name("known_words")
                               .help("Define ratio of known words to be used in the pattern")
                               .required(false)
                               .takes_value(true))
                           .get_matches();

    let time_resolution = Duration::from_secs(value_t!(matches, "time-resolution", u64).unwrap_or(10));
    let logs_per_s = value_t!(matches, "logs-throughput", u64).unwrap_or(0);
    let bytes_per_s = value_t!(matches, "throughput", u64).unwrap_or(0);
    let total_logs = value_t!(matches, "total-logs", u64).unwrap_or(0);
    let path = value_t!(matches, "path", String).unwrap_or_else(|_| "/dev/stdout".to_string());
    let verbose = value_t!(matches, "verbose", bool).unwrap_or_else(true);
    let duration = value_t!(matches, "duration", u64).unwrap_or(0);
    let duration = Duration::from_secs(duration);
    let no_duration = Duration::from_secs(0);

    // Open file to write
    let f = OpenOptions::new().append(true).create(true).open(&path);

    let fd = match f {
        Ok(val) => Some(val),
        Err(err) => {
            eprintln!("Error while opening file to write: {}", err);
            process::exit(1);
        }
    };

    // Read words from dictionary file
    let wordlist = value_t!(matches, "wordlist", String).unwrap_or_else(|_| "/usr/local/wordlist.txt".to_string());
    let wordlist = read_wordlist(&wordlist);
    print(
        verbose,
        format!("[s] {} words read from dictionary!", wordlist.len()),
    );

    // Summarise configuration
    print(
        verbose,
        format!(
            "[s] Going to generate logs into {} with average {} lps and average {} Bps with time resolution: {}s",
            &path,
            logs_per_s,
            bytes_per_s,
            time_resolution.as_secs(),
        )
    );

    // Prepare list of patterns
    // if pattern-file is specified, patterns are get from it
    // otherwise pattern is generated using known_words, random_words, random digits, min and max
    //  - known_words, random_words and random_digits represents ratio eg, 4:5:1, means that 40% of pattern should be known_words etc
    //  - min and max limits length of the pattern in terms of words
    let pattern = value_t!(matches, "pattern", String).unwrap_or_else(|_| "".to_string());
    let random_patterns = value_t!(matches, "random-patterns", u32).unwrap_or(0);
    let pattern_file = value_t!(matches, "pattern-file", String).unwrap_or_else(|_| "".to_string());
    let min = value_t!(matches, "min", u32).unwrap_or(5);
    let max = value_t!(matches, "max", u32).unwrap_or(20);
    let random_words = value_t!(matches, "random_words", u32).unwrap_or(2);
    let random_digits = value_t!(matches, "random_digits", u32).unwrap_or(1);
    let known_words = value_t!(matches, "known_words", u32).unwrap_or(7);

    let patterns: Vec<String> = match pattern_file.as_ref() {
        "" => collect_patterns(
            pattern,
            random_patterns,
            &wordlist,
            min,
            max,
            known_words,
            random_words,
            random_digits,
        ),
        x => read_patterns(x.to_string()),
    };

    for pattern in &patterns {
        print(verbose, format!("Pattern: {}", pattern));
    }

    let start = Instant::now();
    let mut now = Instant::now();
    let mut count_logs = 0;
    let mut count_bytes = 0;
    let mut current_logs = 0;
    let mut current_bytes = 0;
    let mut counter = 0;
    loop {
        // Print statistics
        if now.elapsed() >= time_resolution {
            print(
                verbose,
                format!(
                    "{} pps\t {} b/s",
                    current_logs / now.elapsed().as_secs(),
                    current_bytes / now.elapsed().as_secs()
                ),
            );
            print(
                verbose,
                format!("Total stats: {} logs, {} bytes", count_logs, count_bytes),
            );
            now = Instant::now();
            current_logs = 0;
            current_bytes = 0;
        }

        // Skip iteration because limit of logs/s already reached
        if logs_per_s > 0 && current_logs >= logs_per_s * time_resolution.as_secs() {
            thread::sleep(time_resolution.sub(now.elapsed()));
            continue;
        }

        // Skip iteration because limit of bytes/s already reached
        if bytes_per_s > 0 && current_bytes >= bytes_per_s * time_resolution.as_secs() {
            thread::sleep(time_resolution.sub(now.elapsed()));
            continue;
        }

        // Stop generation if configured duration reached
        if duration > no_duration && now - start > duration {
            print(
                verbose,
                format!(
                    "Logs generation finished after {} seconds",
                    (now - start).as_secs()
                ),
            );
            break;
        }

        // Stop generation if configured limit reached
        if total_logs != 0 && count_logs >= total_logs {
            break;
        }

        // Generate logs from patterns
        for pattern in &patterns {
            let log = build_log(pattern, &wordlist, &mut counter);
            let mut saved_logs: u64 = 0;

            match &fd {
                Some(x) => saved_logs = save_log(x, &log),
                None => {}
            }

            if saved_logs == 1 {
                current_logs += 1;
                current_bytes += log.len() as u64;
                count_logs += 1;
                count_bytes += log.len() as u64;
            }

            if total_logs != 0 && count_logs >= total_logs {
                break;
            }
        }
    }
    print(verbose, format!("Sent {} logs in total", count_logs));
    print(verbose, format!("Sent {} bytes in total", count_bytes));
}

fn save_log(mut fd: &File, log: &str) -> u64 {
    let write = fd.write_all(log.as_bytes());
    match write {
        Ok(_result) => {
            1
        }
        Err(_err) => {
            0
        }
    }
}

fn read_wordlist(filename: &str) -> Vec<String> {
    let mut vec = Vec::<String>::new();
    let file = File::open(filename).unwrap();
    for line in BufReader::new(file).lines() {
        vec.push(line.unwrap());
    }
    vec
}

fn build_log(pattern: &str, wordlist: &[String], counter: &mut u64) -> String {
    let mut rng = rand::thread_rng();
    let slices = pattern.split_whitespace();
    let mut log = "".to_owned();

    for slice in slices {
        if slice.starts_with('{') && slice.ends_with('}') {
            // replace {w} with random word
            if slice.contains('w') {
                log += get_random_word(wordlist);
            }
            // replace {d} with random digit
            else if slice.contains('d') {
                log += &rng.gen_range(0, 0xffffff).to_string();
            }
            // replace {c} with counter value
            else if slice.contains('c') {
                // counter
                log += &counter.to_string();
                *counter += 1;
            }
        } else {
            log += slice;
        }
        log += " ";
    }
    log += "\n";
    log
}

fn get_random_word(wordlist: &[String]) -> &String {
    let mut rng = rand::thread_rng();
    return wordlist.get(rng.gen_range(0, wordlist.len())).unwrap();
}

fn generate_pattern(
    wordlist: &[String],
    min: u32,
    max: u32,
    known_words: u32,
    random_words: u32,
    random_digits: u32,
) -> String {
    let mut rng = rand::thread_rng();
    let mut pattern = "{c} : ".to_string();
    let mut possible_slices = Vec::<String>::new();

    for _ in 0..known_words {
        possible_slices.push(get_random_word(wordlist).to_string());
    }

    for _ in 0..random_words {
        possible_slices.push("{w}".to_string());
    }

    for _ in 0..random_digits {
        possible_slices.push("{d}".to_string());
    }

    let pattern_slices = rng.gen_range(min, max + 1);

    for _ in 0..pattern_slices {
        let position = rng.gen_range(0, possible_slices.len());
        let mut current_pattern = possible_slices.remove(position);
        pattern += &current_pattern;

        if current_pattern != "{d}" && current_pattern != "{w}" {
            current_pattern = get_random_word(wordlist).to_string();
        }

        possible_slices.push(current_pattern);
        pattern += " ";
    }

    pattern
}

// Collect patterns from argument and merge with random patterns
fn collect_patterns(
    pattern: String,
    random_patterns: u32,
    wordlist: &[String],
    min: u32,
    max: u32,
    known_words: u32,
    random_words: u32,
    random_digits: u32,
) -> Vec<String> {
    let mut patterns = Vec::<String>::new();

    if !pattern.eq("") {
        let _patterns = pattern.split('$');
        for pattern in _patterns {
            patterns.push(pattern.to_string());
        }
    }

    for _ in 0..random_patterns {
        patterns.push(generate_pattern(
            wordlist,
            min,
            max,
            known_words,
            random_words,
            random_digits,
        ));
    }

    patterns
}

// Read patterns from file
fn read_patterns(filename: String) -> Vec<String> {
    let mut patterns = Vec::<String>::new();

    let file = File::open(filename).unwrap();
    for line in BufReader::new(file).lines() {
        patterns.push(line.unwrap());
    }

    patterns
}

// Prints message to stderr if verbose is true
fn print(verbose: bool, message: String) {
    if verbose {
        eprintln!("{}", message);
    }
}
