use std::io::BufWriter;

fn main() {
    let stdout = ::std::io::stdout();
    let message = String::from("Hello fellow Rustaceans!");
    let width = message.chars().count();

    let mut writer = BufWriter::new(stdout.lock());
    ::ferris_says::say(&message, width, &mut writer).unwrap();
}
