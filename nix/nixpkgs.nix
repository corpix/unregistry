## Use nixpkgs from environment
#import <nixpkgs>

## Use frozen nixpkgs from internet
with builtins;
import (fetchTarball {
  url    = "https://github.com/nixos/nixpkgs/archive/3f50543e347c074f2834c0a899d3b1fbd626375f.tar.gz";
  sha256 = "1sg97vrihl2n9sh9fvplahkgkqdb4359i95h9515h2fqg6i7gvxp";
})
