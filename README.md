# GoName

GoName is a CLI tool that helps you renaming your video files, inpired by [Terraform](https://github.com/hashicorp/terraform).

## Features

- Rename a large amount of videos files
- Fetch information from TheMovieDB, TheTVDB or AniDB
- Reverting changes

## Usage

Run `goname plan --dir .`

Take a look at the results, if everything looks good, then you can `goname apply --dir . --auto-approve`

You can revert if you want : `goname state revert --all`
