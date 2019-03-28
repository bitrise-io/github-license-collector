require 'bundler'

path = ARGV[0]

base = File.dirname(path)
Dir.chdir(base)

begin
  file = File.open(path, 'r')
rescue => exception
  puts exception
  exit 1
end

# puts file.read()

begin
  bundle = Bundler::LockfileParser.new(Bundler.read_file(file))
rescue => exception
  puts exception
  exit 1
end

gem_name_version_map = bundle.specs.map { |spec|
  [
    spec.name,
    spec.version.to_s,
  ]
}

STDOUT.puts (gem_name_version_map.map { |pair| pair.join(' ') })
