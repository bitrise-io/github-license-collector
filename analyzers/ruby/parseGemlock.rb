require "bundler"

path = $*[0]
# puts path

base = File.dirname(path)
Dir.chdir(base)
file = File.open(path, "r")
# puts file.read()
bundle = Bundler::LockfileParser.new(Bundler.read_file(file))

gem_name_version_map = bundle.specs.map { |spec|
  [
    spec.name,
    # spec.version.to_s,
  ]
}

STDOUT.puts gem_name_version_map.map { |pair| pair.join(" ") }