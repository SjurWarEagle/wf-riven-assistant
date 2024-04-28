# WF Riven Assistant

## Purpose

## Security

### Virus?

I'm aware that there are rare cases of findings of `Trojan.Malware.300983.susgen` or `TrojanDownloader.Agent.gfcf` as
visible in this
report: [VirusTotal](https://www.virustotal.com/gui/file/a13603fed439b2122aa58f87f0c0e29c22f88758596aaa05c49ace63cd9facce?nocache=1)

This is a false positive, a detection that is incorrect: [FAQ](https://go.dev/doc/faq#virus)

### Privacy

This tool does not require any accounts, also it does not read any local files aside from it's configuration
file `config.json`. It only communicates with warframe-market[1] no data is sent to anyone else.

## Confirmation

The configuration is done with the file `config.json` in the tool-folder.

### Example

```json lines
{
  "setup": {
    "platform": "pc",
    "lowerSectionAverageHighlightThreshold": 30
  },
  "rivens": [
    {
      "weapon": "Aeolak",
      "attributes": ""
    }
  ]
}
```

### Parameters

| name                                  |                                             description                                              | possible values    |
|:--------------------------------------|:----------------------------------------------------------------------------------------------------:|:-------------------|
| platform                              |                          defines which system the prices are collected for.                          | pc,ps4,switch,xbox |
| lowerSectionAverageHighlightThreshold |                                                 TODO                                                 | positive number    |
| rivens                                |                                                 TODO                                                 | TODO               |
| weapon                                |                                    the English name of the weapon                                    | Aeolak             |
| attributes                            | This field is currently not supported, but it will be the "riven-suffix" to get more precise results | Igni-decitox       |               |

## Development Notes

### Text?! Really?

**Yes**, not everything needs to be a GUI, good old TUI is good enough for most info.

### Slow auction data collection

The prices for the rives are collected one by one, the requests to warframe-market[1] are done one by one, not in
parallel. Reason is, that warframe-market does not allow this, if more than 1 request at a time is done the request
fails.

### Manual maintenance

Yes you have to maintain the list of rivens by hand, this is done to get the data collection working.
I know it is not comfortable, nor optimal, but it is all I can do at the moment´.

I know how alecaframe[2] is doing it, but I lack the knowledge to use the same process - sorry.
Also, I checked the local files of alecaframe[2] no idea where the riven-info is stored.


[1]: https://warframe.market/

[2]: https://alecaframe.com/