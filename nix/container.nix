{ pkgs        ? import <nixpkgs> {}
, tag         ? version
, pkg         ? pkgs.callPackage ./default.nix { inherit pkgs; }
, tools       ? pkgs.callPackage ./tools.nix   { inherit pkgs; }
, name
, namespace
, version
, ... }:
with builtins;
with pkgs;
with lib;
let
  contents = [
    cacert coreutils busybox
    curl iproute bashInteractive
    pkg
  ]
  ++ tools.mkNss { }
  ++ tools.mkHosts { }
  ++ tools.mkUsers { users = [{ name = "nobody"; uid = 15100; gid = 15100; }]; };
  timeZone = "UTC";
in dockerTools.buildLayeredImage {
  inherit tag;

  name = "${namespace}/${name}";

  inherit contents;

  # https://github.com/moby/moby/blob/master/image/spec/v1.2.md
  config = {
    User = "nobody";
    Env = ["TZ=${timeZone}"];
    Entrypoint = ["/bin/${name}" "-c" "/etc/${name}/config.yml"];
    Expose = ["4180/tcp" "4280/tcp"];
  };
}
