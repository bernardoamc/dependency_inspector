# Usage: PAT=<github_personal_access_token> ORG=<organization_name> ruby scripts/fetch_github_lock_files.rb <filename>
# Set PUBLIC_REPO to true to fetch only from public repositories in the search.
#
# This script fetches a file from every repository in the organization and saves them in the OUTPUT_DIR directory.
# It will skip repositories that already have the file.
#
# Each file will be renamed to the repository name and saved with the .lock extension.

require 'octokit'
require 'net/http'

OUTPUT_DIR = 'input'.freeze

access_token = ENV['PAT']
organization_name = ENV['ORG']
public_repo = ENV.fetch('PUBLIC_REPO', false)
filename = ARGV[0]

if access_token.nil? || organization_name.nil?
  puts 'Error: Please set the PAT and ORG environment variables.'
  exit 1
end

def fetch_file(client, repo)
  repo_name = repo.name.gsub('-', '_')

  if File.exist?("#{OUTPUT_DIR}/#{repo_name}.lock")
    puts "#{filename} already exists for #{repo.name}"
    return
  end

  file_content = client.contents(repo.full_name, path: filename, accept: 'application/vnd.github.v3.raw')

  File.open("#{GEMFILE_DIR}/#{repo_name}.lock", 'wb') do |file|
    file.write(file_content)
  end

  puts "Fetched #{filename} for #{repo.name}"
rescue Octokit::TooManyRequests => e
  puts "Rate limit exceeded. Sleeping for #{e[:response_headers][:retry_after]} seconds."
  sleep e[:response_headers][:retry_after].to_i
  retry
rescue Octokit::NotFound
  puts "#{filename} not found for #{repo.name}"
end

client = Octokit::Client.new(access_token: access_token)
page = 1

begin
  loop do
    options = { per_page: 30, page: page }
    options[:type] = 'public' if public_repo
    repositories = client.organization_repositories(organization_name, options)
    break if repositories.empty?

    repositories.each do |repo|
      next if repo.archived?
      fetch_file(client, repo)
    end

    page += 1
  end
rescue Octokit::Error => e
  puts "An error occurred: #{e.message}"
end
