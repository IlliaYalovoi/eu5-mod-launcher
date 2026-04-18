Feature request:

In mod launcher I want to see is a mod compatible with the current version of the game or not.

Important: mod "version" != mod "supported_game_version". Only latter we should look onto.

"Supported game version" is a field in mod manifest, which should be filled by mod author. It can be a single version (e.g. "1.0.0") or a string that supports multiple versions with asterisk formatting (e.g. "1.0.*" or "1.*").

How to identify current game version? That is where it becomes tricky. Paradox games was created in different times and uses different ways to store version information.

Crusader Kings 3 - 
    - filename: clausewitz_branch.txt - content: titus/release/1.18.2
    - filename: titus_branch.txt - content: release/1.18.4
    - Both files is exist at the same time, game version is 1.18.4 at time of writing. That means that titus_branch.txt should be source of truth, but if it is missing, we can fallback to clausewitz_branch.txt. If both files is missing (or have broken formatting), we can assume that game version is unknown.

Europa Universalis 4 -
    - filename: clausewitz_branch.txt - content: release_1.37.5
    - filename: eu4branch.txt - content: release_1.37.5
    - Here is seems aligned, both files have the same content, but we can assume that eu4branch.txt is source of truth, and if it is missing, we can fallback to clausewitz_branch.txt. If both files is missing (or have broken formatting), we can assume that game version is unknown.

Europa Universalis 5 -
    - filename: clausewitz_branch.txt - content: caesar/release/1.1.0
    - filename: caesar_branch.txt - content: release/1.1.0
    - Main problem here, is that current patch is 1.1.10, so both files is outdated. I think we should also give user an ability to manually set game version (in borders of formatting, of course) and store it locally (with ability to auto-detect it again, if user wants to). Auto detect should fallback firstly to ceasar_branch.txt, then to clausewitz_branch.txt, and if both files is missing (or have broken formatting), we can assume that game version is unknown.

Hearts of Iron 4 - 
    - filename: clausewitz_branch.txt - content: None
    - filename: ho4branch.txt - content: None
    - That one even worse, HOI4 does not stores any version information in files at all (both files have simple "None" written in them), so user selection will be manual flow. In UI, when we can not identify version, we should show "Unknown version" and soft warning/reminder (not aggressive but visible) to user that they can set it manually in settings, and that it is required for correct mod compatibility check.

Victoria 3 - 
    - filename: clausewitz_branch.txt - content: caligula/release/1.12.5
    - filename: caligula_branch.txt - content: release/1.12.5
    - That one is similar to EU5, but at least files is up to date. Follow usual pattern 