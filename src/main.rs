extern crate crypto;
use std::path::Path;

fn main() {
	let bios: [u8; 256] = read_bios(Path::new("DMG_ROM.gb"));
}

fn read_bios(path: &Path) -> [u8; 256] {
	use crypto::digest::Digest;
	use crypto::sha2::Sha256;
	use std::fs::File;
	use std::io::Read;

	let mut file: File = File::open(path).unwrap();

	// We verify that the file is 256 bytes long.
	let file_size: u64 = file.metadata().unwrap().len();
	assert!(file_size == 256);

	let mut bios: [u8; 256] = [0x00; 256];
	file.read_exact(&mut bios).unwrap();

	// We verify the SHA-256 hash.
	let mut hasher: Sha256 = Sha256::new();
	hasher.input(&bios);
	let hash: String = hasher.result_str();
	assert!(hash == "cf053eccb4ccafff9e67339d4e78e98dce7d1ed59be819d2a1ba2232c6fce1c7");

	bios
}