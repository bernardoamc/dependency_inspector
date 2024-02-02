# Usage: ruby public_gems.rb <registry_file_path>
#
# This script checks the gems in the registry file and compares them with the gems in the rubygems registry.
# It then prints:
#
#   1. the number of gems living in rubygems that are present in the registry file
#   2. the number of gems with the same name as the ones in the registry file that are not claimed by the GEM_AUTHOR_REGEX in rubygems
#   3. the number of gems from the registry file that are missing from rubygems
#   4. the different kinds of authors in rubygems from gems that are present (e.g. acme, acme inc, acme ltd)

require 'net/http'
require 'json'

GEM_AUTHOR_REGEX = /acme/.freeze

def check_gem_author(gem_name)
  url = URI("https://rubygems.org/api/v1/gems/#{gem_name}.json")
  response = Net::HTTP.get_response(url)

  if response.is_a?(Net::HTTPSuccess)
    begin
      gem_info = JSON.parse(response.body)
      [:ok, gem_name, gem_info['authors']]
    rescue JSON::ParserError
      [:parser_error, gem_name, nil]
    end
  else
    [:not_found, gem_name, nil]
  end
end

registry_path = ARGV[0] || 'registry.json'
gems = JSON.parse(File.read(registry_path)).fetch('dependencies')
registered = {}
missing = []

gems.each do |gem|
  puts "Checking #{gem}..."

  case check_gem_author(gem)
    in [:ok, gem_name, authors]
      registered[gem_name] = authors
    in [:not_found, gem_name, _]
      missing << gem_name
    in [:parser_error, gem_name, _]
      puts "Error parsing JSON for #{gem_name}"
  end
end

buckets = registered.each_with_object(Hash.new { |h, k| h[k] = [] }) do |(gem, author), buckets|
  if author.downcase =~ GEM_AUTHOR_REGEX
    buckets[:org] << gem
  else
    buckets[:other] << gem
  end
end

puts buckets.inspect

authors = buckets[:org].map { |gem| registered[gem] }.uniq.inspect

puts "Gems in our private registry that are claimed by us in rubygems.org: #{buckets[:org].size}"
puts "Gems in our private registry that are NOT claimed by us in rubygems.org: #{buckets[:other].size}"
puts "Gems in our private registry that are missing in rubygems.org: #{missing.size}"
puts "Different kinds of authors from our organization in rubygems.org: #{authors.inspect}"
