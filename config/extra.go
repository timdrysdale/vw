package config

//ExtrasConfiguration struct is for non-essential values
//
//Variables is provided as a flexible resource for configuration writers,
//to assist in reducing copy-paste mistakes. Commonly occuring parts of paths
//and command lines can be specified as pseudo-environment variables
//that are scoped to within the executable. For the case where identical
//environment variable(s) exist, behaviour is currently undefined. Some
//variables are reserved, e.g. ${localport}, and are ignored.
//
//Example YAML showing usage of variables
// ---
// variables:
//   -   relay: "wss://video.practable.io:443/in"
//       uuid: "xxx-000-yyy"
//       session: "aaa-111-bbb"
// stream:
//   -   to: "${relay}/in/${uuid}/${session}/front/medium"
//  	 serve: "${localport}/out/front/medium"
//       from:
//         - audio
//         - videoFrontMedium
type ExtraConfiguration struct {
	Variables map[string]string `mapstructure:"variables"`
}
