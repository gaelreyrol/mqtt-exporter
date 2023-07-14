{ lib
, buildGoModule
}:

buildGoModule {
  pname = "mqtt-exporter";
  version = "dev";

  src = ./.;

  vendorHash = "sha256-SA2sjZfisHLpDm1820GToerHLbE1oQ2obl9pmsiyRqE=";

  ldflags = [
    "-s"
    "-w"
  ];

  meta = with lib; {
    description = "Export MQTT messages to Prometheus";
    homepage = "https://github.com/gaelreyrol/mqtt-exporter";
    license = licenses.mit;
    maintainers = with maintainers; [ gaelreyrol ];
    mainProgram = "mqtt_exporter";
  };
}
