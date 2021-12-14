{ config, lib, pkgs, ... }:
with builtins;
with lib;

let
  name = "unregistry";
  cfg = config.services."${name}";
  pkg = pkgs.callPackage ./default.nix { };
in {
  options = with types; {
    services."${name}" = {
      enable = mkEnableOption "Goboilerplate";

      user = mkOption {
        default = name;
        type = str;
        description = "User name to run service from";
      };
      group = mkOption {
        default = name;
        type = str;
        description = "Group name to run service from";
      };

      config = mkOption {
        type = attrs;
        default = { };
        description = "Goboilerplate raw configuration";
      };
    };
  };

  config = optionalAttrs cfg.enable {
    users = {
      extraUsers = mkIf (name == cfg.user)
        {
          ${name} = {
            name = cfg.user;
            group = cfg.group;
          };
        };

      extraGroups = optionalAttrs (name == cfg.group)
        { ${name}.name = cfg.group; };
    };

    systemd.services.${name} = {
      enable = true;
      wantedBy = ["multi-user.target"];

      serviceConfig = {
        Type       = "simple";
        Restart    = "on-failure";
        RestartSec = 1;

        User  = cfg.user;
        Group = cfg.group;

        ExecStart = "${pkg}/bin/${name} -c ${pkgs.writeText "config.yml" (toJSON cfg.config)}";
      };
    };
  };
}
