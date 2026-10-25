import fs from 'node:fs';
import { validateVersion } from './release-version.mjs';
const manifest = JSON.parse(fs.readFileSync(process.argv[2], 'utf8'));
const version = validateVersion(manifest.version);
const base = `https://github.com/MohamedElashri/snipo/releases/download/v${version}`;
function platform(os) {
  return `  on_${os === 'darwin' ? 'macos' : 'linux'} do\n` + ['arm64', 'amd64'].map(arch => {
    const name = `snippy_${version}_${os}_${arch}.tar.gz`;
    const asset = manifest.assets.find(a => a.name === name);
    if (!asset || !/^[a-f0-9]{64}$/.test(asset.sha256)) throw new Error(`Missing hash for ${name}`);
    return `    on_${arch === 'arm64' ? 'arm' : 'intel'} do\n      url "${base}/${name}"\n      sha256 "${asset.sha256}"\n    end\n`;
  }).join('') + '  end\n';
}
console.log(`class Snippy < Formula
  desc "Terminal client for the Snipo snippet manager"
  homepage "https://github.com/MohamedElashri/snipo"
  version "${version}"
  license "AGPL-3.0-only"

${platform('darwin')}
${platform('linux')}
  def install
    bin.install "snippy"
  end

  test do
    assert_match "${version}", shell_output("#{bin}/snippy version")
  end
end`);
