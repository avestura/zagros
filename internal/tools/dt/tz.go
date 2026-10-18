package dt

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
	_ "time/tzdata" // Windows has no system zoneinfo: embed the IANA database
)

// zoneNames is a static list of IANA names: Go cannot enumerate the embedded
// database, so the list is kept here (tests load every entry).
var zoneNames = strings.Fields(`
UTC GMT
Africa/Abidjan Africa/Accra Africa/Addis_Ababa Africa/Algiers Africa/Asmara Africa/Bamako Africa/Bangui Africa/Banjul Africa/Bissau
Africa/Blantyre Africa/Brazzaville Africa/Bujumbura Africa/Cairo Africa/Casablanca Africa/Ceuta Africa/Conakry Africa/Dakar
Africa/Dar_es_Salaam Africa/Djibouti Africa/Douala Africa/El_Aaiun Africa/Freetown Africa/Gaborone Africa/Harare Africa/Johannesburg
Africa/Juba Africa/Kampala Africa/Khartoum Africa/Kigali Africa/Kinshasa Africa/Lagos Africa/Libreville Africa/Lome Africa/Luanda
Africa/Lubumbashi Africa/Lusaka Africa/Malabo Africa/Maputo Africa/Maseru Africa/Mbabane Africa/Mogadishu Africa/Monrovia
Africa/Nairobi Africa/Ndjamena Africa/Niamey Africa/Nouakchott Africa/Ouagadougou Africa/Porto-Novo Africa/Sao_Tome Africa/Tripoli
Africa/Tunis Africa/Windhoek
America/Adak America/Anchorage America/Anguilla America/Antigua America/Araguaina America/Argentina/Buenos_Aires
America/Argentina/Cordoba America/Argentina/Mendoza America/Argentina/Ushuaia America/Aruba America/Asuncion America/Atikokan
America/Bahia America/Bahia_Banderas America/Barbados America/Belem America/Belize America/Blanc-Sablon America/Boa_Vista
America/Bogota America/Boise America/Cambridge_Bay America/Campo_Grande America/Cancun America/Caracas America/Cayenne
America/Cayman America/Chicago America/Chihuahua America/Costa_Rica America/Creston America/Cuiaba America/Curacao
America/Danmarkshavn America/Dawson America/Dawson_Creek America/Denver America/Detroit America/Dominica America/Edmonton
America/Eirunepe America/El_Salvador America/Fort_Nelson America/Fortaleza America/Glace_Bay America/Goose_Bay America/Grand_Turk
America/Grenada America/Guadeloupe America/Guatemala America/Guayaquil America/Guyana America/Halifax America/Havana
America/Hermosillo America/Indiana/Indianapolis America/Inuvik America/Iqaluit America/Jamaica America/Juneau
America/Kentucky/Louisville America/Kralendijk America/La_Paz America/Lima America/Los_Angeles America/Lower_Princes
America/Maceio America/Managua America/Manaus America/Marigot America/Martinique America/Matamoros America/Mazatlan
America/Menominee America/Merida America/Metlakatla America/Mexico_City America/Miquelon America/Moncton America/Monterrey
America/Montevideo America/Montserrat America/Nassau America/New_York America/Nome America/Noronha America/North_Dakota/Center
America/Nuuk America/Ojinaga America/Panama America/Paramaribo America/Phoenix America/Port-au-Prince America/Port_of_Spain
America/Porto_Velho America/Puerto_Rico America/Punta_Arenas America/Rankin_Inlet America/Recife America/Regina America/Resolute
America/Rio_Branco America/Santarem America/Santiago America/Santo_Domingo America/Sao_Paulo America/Scoresbysund America/Sitka
America/St_Barthelemy America/St_Johns America/St_Kitts America/St_Lucia America/St_Thomas America/St_Vincent America/Swift_Current
America/Tegucigalpa America/Thule America/Tijuana America/Toronto America/Tortola America/Vancouver America/Whitehorse
America/Winnipeg America/Yakutat
Antarctica/Casey Antarctica/Davis Antarctica/DumontDUrville Antarctica/Macquarie Antarctica/Mawson Antarctica/McMurdo
Antarctica/Palmer Antarctica/Rothera Antarctica/Syowa Antarctica/Troll Antarctica/Vostok Arctic/Longyearbyen
Asia/Aden Asia/Almaty Asia/Amman Asia/Anadyr Asia/Aqtau Asia/Aqtobe Asia/Ashgabat Asia/Atyrau Asia/Baghdad Asia/Bahrain Asia/Baku
Asia/Bangkok Asia/Barnaul Asia/Beirut Asia/Bishkek Asia/Brunei Asia/Chita Asia/Colombo Asia/Damascus Asia/Dhaka Asia/Dili
Asia/Dubai Asia/Dushanbe Asia/Famagusta Asia/Gaza Asia/Hebron Asia/Ho_Chi_Minh Asia/Hong_Kong Asia/Hovd Asia/Irkutsk
Asia/Jakarta Asia/Jayapura Asia/Jerusalem Asia/Kabul Asia/Kamchatka Asia/Karachi Asia/Kathmandu Asia/Khandyga Asia/Kolkata
Asia/Krasnoyarsk Asia/Kuala_Lumpur Asia/Kuching Asia/Kuwait Asia/Macau Asia/Magadan Asia/Makassar Asia/Manila Asia/Muscat
Asia/Nicosia Asia/Novokuznetsk Asia/Novosibirsk Asia/Omsk Asia/Oral Asia/Phnom_Penh Asia/Pontianak Asia/Pyongyang Asia/Qatar
Asia/Qostanay Asia/Qyzylorda Asia/Riyadh Asia/Sakhalin Asia/Samarkand Asia/Seoul Asia/Shanghai Asia/Singapore Asia/Srednekolymsk
Asia/Taipei Asia/Tashkent Asia/Tbilisi Asia/Tehran Asia/Thimphu Asia/Tokyo Asia/Tomsk Asia/Ulaanbaatar Asia/Urumqi
Asia/Ust-Nera Asia/Vientiane Asia/Vladivostok Asia/Yakutsk Asia/Yangon Asia/Yekaterinburg Asia/Yerevan
Atlantic/Azores Atlantic/Bermuda Atlantic/Canary Atlantic/Cape_Verde Atlantic/Faroe Atlantic/Madeira Atlantic/Reykjavik
Atlantic/South_Georgia Atlantic/St_Helena Atlantic/Stanley
Australia/Adelaide Australia/Brisbane Australia/Broken_Hill Australia/Darwin Australia/Eucla Australia/Hobart Australia/Lindeman
Australia/Lord_Howe Australia/Melbourne Australia/Perth Australia/Sydney
Europe/Amsterdam Europe/Andorra Europe/Astrakhan Europe/Athens Europe/Belgrade Europe/Berlin Europe/Bratislava Europe/Brussels
Europe/Bucharest Europe/Budapest Europe/Busingen Europe/Chisinau Europe/Copenhagen Europe/Dublin Europe/Gibraltar Europe/Guernsey
Europe/Helsinki Europe/Isle_of_Man Europe/Istanbul Europe/Jersey Europe/Kaliningrad Europe/Kirov Europe/Kyiv Europe/Lisbon
Europe/Ljubljana Europe/London Europe/Luxembourg Europe/Madrid Europe/Malta Europe/Mariehamn Europe/Minsk Europe/Monaco
Europe/Moscow Europe/Oslo Europe/Paris Europe/Podgorica Europe/Prague Europe/Riga Europe/Rome Europe/Samara Europe/San_Marino
Europe/Sarajevo Europe/Saratov Europe/Simferopol Europe/Skopje Europe/Sofia Europe/Stockholm Europe/Tallinn Europe/Tirane
Europe/Ulyanovsk Europe/Vaduz Europe/Vatican Europe/Vienna Europe/Vilnius Europe/Volgograd Europe/Warsaw Europe/Zagreb Europe/Zurich
Indian/Antananarivo Indian/Chagos Indian/Christmas Indian/Cocos Indian/Comoro Indian/Kerguelen Indian/Mahe Indian/Maldives
Indian/Mauritius Indian/Mayotte Indian/Reunion
Pacific/Apia Pacific/Auckland Pacific/Bougainville Pacific/Chatham Pacific/Chuuk Pacific/Easter Pacific/Efate Pacific/Fakaofo
Pacific/Fiji Pacific/Funafuti Pacific/Galapagos Pacific/Gambier Pacific/Guadalcanal Pacific/Guam Pacific/Honolulu
Pacific/Kanton Pacific/Kiritimati Pacific/Kosrae Pacific/Kwajalein Pacific/Majuro Pacific/Marquesas Pacific/Midway Pacific/Nauru
Pacific/Niue Pacific/Norfolk Pacific/Noumea Pacific/Pago_Pago Pacific/Palau Pacific/Pitcairn Pacific/Pohnpei Pacific/Port_Moresby
Pacific/Rarotonga Pacific/Saipan Pacific/Tahiti Pacific/Tarawa Pacific/Tongatapu Pacific/Wake Pacific/Wallis
US/Alaska US/Arizona US/Central US/Eastern US/Hawaii US/Mountain US/Pacific
Canada/Atlantic Canada/Central Canada/Eastern Canada/Mountain Canada/Newfoundland Canada/Pacific
Asia/Calcutta Asia/Katmandu Asia/Saigon Asia/Rangoon Asia/Istanbul Asia/Tel_Aviv Asia/Dacca Europe/Kiev America/Buenos_Aires
Australia/ACT Australia/Canberra Australia/NSW Australia/Queensland Australia/Victoria Australia/West
Etc/UTC Etc/GMT Etc/GMT+1 Etc/GMT+2 Etc/GMT+3 Etc/GMT+4 Etc/GMT+5 Etc/GMT+6 Etc/GMT+7 Etc/GMT+8 Etc/GMT+9 Etc/GMT+10 Etc/GMT+11 Etc/GMT+12
Etc/GMT-1 Etc/GMT-2 Etc/GMT-3 Etc/GMT-4 Etc/GMT-5 Etc/GMT-6 Etc/GMT-7 Etc/GMT-8 Etc/GMT-9 Etc/GMT-10 Etc/GMT-11 Etc/GMT-12 Etc/GMT-13 Etc/GMT-14
Iran Israel Japan Egypt Turkey Singapore Hongkong Poland Portugal Cuba Jamaica Navajo Eire GB NZ PRC ROC ROK
EST5EDT CST6CDT MST7MDT PST8PDT CET MET EET WET
`)

var (
	zoneIndexOnce sync.Once
	zoneLower     map[string]string
)

func zoneIdx() map[string]string {
	zoneIndexOnce.Do(func() {
		zoneLower = make(map[string]string, len(zoneNames)*2)
		for _, n := range zoneNames {
			zoneLower[strings.ToLower(n)] = n
		}
	})
	return zoneLower
}

// canonicalZone resolves a case-insensitive IANA name or a bare city ("tokyo",
// "new york") to the list entry.
func canonicalZone(name string) (string, bool) {
	k := strings.ToLower(strings.TrimSpace(name))
	idx := zoneIdx()
	if v, ok := idx[k]; ok {
		return v, true
	}
	k2 := strings.NewReplacer(" ", "_").Replace(k)
	if v, ok := idx[k2]; ok {
		return v, true
	}
	if !strings.Contains(k2, "/") && k2 != "" {
		// city shortcut: first list entry whose last segment matches
		for _, n := range zoneNames {
			if strings.HasPrefix(n, "Etc/") || strings.HasPrefix(n, "US/") || strings.HasPrefix(n, "Canada/") {
				continue
			}
			i := strings.LastIndex(n, "/")
			if i >= 0 && strings.ToLower(n[i+1:]) == k2 {
				return n, true
			}
		}
	}
	return "", false
}

type abbr struct {
	off  int // seconds east of UTC
	note string
}

// zoneAbbrs maps common abbreviations to fixed offsets. Ambiguous ones use the
// most common meaning (CST = US Central, IST = India).
var zoneAbbrs = map[string]abbr{
	"UTC": {0, ""}, "GMT": {0, ""}, "UT": {0, ""}, "Z": {0, ""}, "WET": {0, ""}, "WEST": {3600, ""},
	"BST": {3600, "British Summer Time"}, "IST": {19800, "India Standard Time"}, "CET": {3600, ""}, "CEST": {7200, ""},
	"EET": {7200, ""}, "EEST": {10800, ""}, "MSK": {10800, ""}, "SAST": {7200, ""}, "EAT": {10800, ""}, "WAT": {3600, ""}, "CAT": {7200, ""},
	"EST": {-5 * 3600, ""}, "EDT": {-4 * 3600, ""}, "CST": {-6 * 3600, "US Central Standard Time"}, "CDT": {-5 * 3600, ""},
	"MST": {-7 * 3600, ""}, "MDT": {-6 * 3600, ""}, "PST": {-8 * 3600, ""}, "PDT": {-7 * 3600, ""},
	"AKST": {-9 * 3600, ""}, "AKDT": {-8 * 3600, ""}, "HST": {-10 * 3600, ""}, "AST": {-4 * 3600, ""}, "ADT": {-3 * 3600, ""},
	"NST": {-12600, ""}, "NDT": {-9000, ""}, "PKT": {5 * 3600, ""}, "ICT": {7 * 3600, ""}, "WIB": {7 * 3600, ""},
	"SGT": {8 * 3600, ""}, "HKT": {8 * 3600, ""}, "AWST": {8 * 3600, ""}, "JST": {9 * 3600, ""}, "KST": {9 * 3600, ""},
	"ACST": {34200, ""}, "ACDT": {37800, ""}, "AEST": {10 * 3600, ""}, "AEDT": {11 * 3600, ""},
	"NZST": {12 * 3600, ""}, "NZDT": {13 * 3600, ""}, "IRST": {12600, "Iran Standard Time"}, "IRDT": {16200, ""}, "GST": {4 * 3600, ""},
}

var offsetRe = regexp.MustCompile(`(?i)^(?:utc|gmt|ut)?\s*([+\-−])\s*(\d{1,2})(?::?(\d{2}))?(?::?(\d{2}))?$`)

// parseOffset reads "+05:30", "UTC+3", "GMT-5", "-0800".
func parseOffset(s string) (int, bool) {
	m := offsetRe.FindStringSubmatch(strings.TrimSpace(s))
	if m == nil {
		return 0, false
	}
	h, _ := strconv.Atoi(m[2])
	mi, _ := strconv.Atoi(m[3])
	se, _ := strconv.Atoi(m[4])
	if h > 18 || mi > 59 || se > 59 {
		return 0, false
	}
	off := h*3600 + mi*60 + se
	if m[1] == "-" || m[1] == "−" {
		off = -off
	}
	return off, true
}

func offsetName(off int) string {
	sign := '+'
	if off < 0 {
		sign, off = '-', -off
	}
	if off%3600 == 0 {
		return fmt.Sprintf("UTC%c%02d", sign, off/3600)
	}
	return fmt.Sprintf("UTC%c%02d:%02d", sign, off/3600, off%3600/60)
}

// loadLoc resolves a zone name; empty means the local zone.
func loadLoc(name string) (*time.Location, error) {
	n := strings.TrimSpace(name)
	switch strings.ToLower(n) {
	case "", "local", "system":
		return time.Local, nil
	case "utc", "z", "ut":
		return time.UTC, nil
	}
	if a, ok := zoneAbbrs[strings.ToUpper(n)]; ok {
		return time.FixedZone(strings.ToUpper(n), a.off), nil
	}
	if off, ok := parseOffset(n); ok {
		return time.FixedZone(offsetName(off), off), nil
	}
	if c, ok := canonicalZone(n); ok {
		loc, err := time.LoadLocation(c)
		if err == nil {
			return loc, nil
		}
	}
	if loc, err := time.LoadLocation(n); err == nil {
		return loc, nil
	}
	return nil, fmt.Errorf("unknown time zone %q (use an IANA name like Europe/Paris, a city like Tokyo, an abbreviation like EST or an offset like UTC+3:30)", name)
}

func fmtOffset(off int, colon bool) string {
	sign := '+'
	if off < 0 {
		sign, off = '-', -off
	}
	if colon {
		return fmt.Sprintf("%c%02d:%02d", sign, off/3600, off%3600/60)
	}
	return fmt.Sprintf("%c%02d%02d", sign, off/3600, off%3600/60)
}

func filterZones(filter string) []string {
	f := strings.ToLower(strings.NewReplacer(" ", "_").Replace(strings.TrimSpace(filter)))
	var out []string
	for _, n := range zoneNames {
		if f == "" || strings.Contains(strings.ToLower(n), f) {
			out = append(out, n)
		}
	}
	sort.Strings(out)
	return out
}
