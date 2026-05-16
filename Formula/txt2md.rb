class Txt2md < Formula
  desc "Convert plain text to well-formatted Markdown"
  homepage "https://github.com/Somehow007/txt2md"
  url "https://github.com/Somehow007/txt2md/releases/download/v0.1.0/txt2md_Darwin_arm64_v0.1.0.tar.gz"
  version "0.1.0"
  sha256 "PLACEHOLDER_SHA256"

  on_macos do
    on_arm do
      url "https://github.com/Somehow007/txt2md/releases/download/v#{version}/txt2md_Darwin_arm64_v#{version}.tar.gz"
      sha256 "PLACEHOLDER_DARWIN_ARM64_SHA256"
    end
    on_intel do
      url "https://github.com/Somehow007/txt2md/releases/download/v#{version}/txt2md_Darwin_x86_64_v#{version}.tar.gz"
      sha256 "PLACEHOLDER_DARWIN_AMD64_SHA256"
    end
  end

  on_linux do
    on_arm do
      url "https://github.com/Somehow007/txt2md/releases/download/v#{version}/txt2md_Linux_arm64_v#{version}.tar.gz"
      sha256 "PLACEHOLDER_LINUX_ARM64_SHA256"
    end
    on_intel do
      url "https://github.com/Somehow007/txt2md/releases/download/v#{version}/txt2md_Linux_x86_64_v#{version}.tar.gz"
      sha256 "PLACEHOLDER_LINUX_AMD64_SHA256"
    end
  end

  def install
    bin.install "txt2md"
    man1.install "man/txt2md.1" if File.exist?(File.join(buildpath, "man", "txt2md.1"))
    bash_completion.install "completions/txt2md.bash" if File.exist?(File.join(buildpath, "completions", "txt2md.bash"))
    zsh_completion.install "completions/txt2md.zsh" => "_txt2md" if File.exist?(File.join(buildpath, "completions", "txt2md.zsh"))
    fish_completion.install "completions/txt2md.fish" if File.exist?(File.join(buildpath, "completions", "txt2md.fish"))
  end

  test do
    system "#{bin}/txt2md", "--version"
  end
end
