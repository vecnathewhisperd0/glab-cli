require "gitlab-dangerfiles"

Gitlab::Dangerfiles.for_project(self) do |dangerfiles|
  dangerfiles.import_plugins
  dangerfiles.import_dangerfiles(except: %w[commit_messages])
end
