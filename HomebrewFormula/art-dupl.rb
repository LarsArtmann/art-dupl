# typed: false
# frozen_string_literal: true

# Homebrew formula for art-dupl - code duplication detection tool
# Install with: brew install LarsArtmann/art-dupl/art-dupl
class Artdupl < Formula
  desc "Code duplication detection tool using suffix tree algorithms"
  homepage "https://github.com/LarsArtmann/art-dupl"
  version "1.0.0"
  license "MIT"

  on_macos do
    on_intel do
      url "https://github.com/LarsArtmann/art-dupl/releases/download/v#{version}/art-dupl_#{version}_darwin_amd64.tar.gz"
      sha256 "__SHA256_AMD64__"
    end
    on_arm do
      url "https://github.com/LarsArtmann/art-dupl/releases/download/v#{version}/art-dupl_#{version}_darwin_arm64.tar.gz"
      sha256 "__SHA256_ARM64__"
    end
  end

  on_linux do
    on_intel do
      url "https://github.com/LarsArtmann/art-dupl/releases/download/v#{version}/art-dupl_#{version}_linux_amd64.tar.gz"
      sha256 "__SHA256_AMD64__"
    end
    on_arm do
      url "https://github.com/LarsArtmann/art-dupl/releases/download/v#{version}/art-dupl_#{version}_linux_arm64.tar.gz"
      sha256 "__SHA256_ARM64__"
    end
  end

  def install
    bin.install "art-dupl"
    generate_completions_from_executable(bin/"art-dupl", "completion")
  end

  test do
    assert_match "art-dupl version", shell_output("#{bin}/art-dupl --version")
    assert_match "Code Duplication Statistics", shell_output("#{bin}/art-dupl stats --help")
  end
end
