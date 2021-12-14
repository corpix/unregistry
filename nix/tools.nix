{ pkgs ? import ./nixpkgs.nix {} }:
with builtins;
with pkgs;
with lib;
{
  mkNss = { }: [
    (
      writeTextDir "etc/nsswitch.conf" ''
        passwd:    files
        group:     files
        shadow:    files
        hosts:     files dns
      ''
    )
  ];
  mkHosts = { hosts ? [] }: [
    (
      writeTextDir "etc/hosts" ''
        127.0.0.1 localhost
        ::1       localhost

        # user defined hosts:
        ${concatStringsSep "\n" hosts}
      ''
    )
  ];

  mkUsers = { users ? [] }: [
    (
      writeTextDir "etc/shadow" ''
        root:!x:::::::
        ${concatMapStringsSep "\n" (user: "${user.name}:!:::::::") users}
      ''
    )
    (
      writeTextDir "etc/passwd" ''
        root:x:0:0::/root:${runtimeShell}
        ${concatMapStringsSep
          "\n"
          (user: "${user.name}:x:${toString user.uid}:${toString user.gid}::/home/${user.name}:")
          users}
      ''
    )
    (
      writeTextDir "etc/group" ''
        root:x:0:
        ${concatMapStringsSep
          "\n"
          (user: "${user.name}:x:${toString user.gid}:")
          users}
      ''
    )
    (
      writeTextDir "etc/gshadow" ''
        root:x::
        ${concatMapStringsSep
          "\n"
          (user: "${user.name}:x::")
          users}
      ''
    )
  ];

  tagWithRev = { tag, rev }:
    if typeOf rev == "string" && rev != ""
    then "${tag}-${rev}"
    else tag;
}
