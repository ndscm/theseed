"""Export a branch's commit history as a rebase-todo or JSON.

In rebase-todo format, every commit emits two lines:

    pick <sha>
    exec GIT_AUTHOR_NAME=... ... git commit --amend --no-edit

Running `git rebase` with this todo over an orphan root restores each commit's
original author and committer identity and timestamps.

In JSON format, the output is {"commits": [{<metadata>}, ...]}, one entry per
commit, including the list of Signed-off-by trailers (JSON only).
"""

import argparse
import datetime
import json
import shlex
import subprocess
import zoneinfo

_COMMIT_SEP = "\x00"
_FIELD_SEP = "\x1f"
_SIGNOFF_SEP = "\x1e"
_FORMAT = _FIELD_SEP.join(
    [
        "%H",
        "%an",
        "%ae",
        "%aI",
        "%cn",
        "%ce",
        "%cI",
        f"%(trailers:key=Signed-off-by,valueonly=true,unfold=true,separator={_SIGNOFF_SEP})",
    ]
)


def read_commits(branch: str, *, worktree: str = ".") -> list[dict[str, object]]:
    process = subprocess.run(
        ["git", "log", "--reverse", "-z", "--format=" + _FORMAT, branch],
        cwd=worktree,
        stdout=subprocess.PIPE,
        stderr=subprocess.DEVNULL,
    )
    if process.returncode:
        return []
    commits = []
    for record in process.stdout.decode().split(_COMMIT_SEP):
        if not record:
            continue
        fields = record.split(_FIELD_SEP)
        commits.append(
            {
                "commit": fields[0],
                "authorName": fields[1],
                "authorEmail": fields[2],
                "authorDate": fields[3],
                "committerName": fields[4],
                "committerEmail": fields[5],
                "committerDate": fields[6],
                "signedOffBy": fields[7].split(_SIGNOFF_SEP) if fields[7] else [],
            }
        )
    return commits


def to_san_francisco_time(iso: str) -> str:
    parsed = datetime.datetime.fromisoformat(iso)
    return parsed.astimezone(zoneinfo.ZoneInfo("America/Los_Angeles")).isoformat()


BAD_COMMITS = [
    "9e7c95051cc08aecb69207168ccd657816f5697d",
    "ff5e54a47a72eb6d7bf11319e9120b0b97488201",
    "b91efba33ec1ae697094a3a827f4585d4c9ed7d2",
    "0636917ad52d23ae27c897b99501646c86037b06",
    "897ec414bbfe24722b63a709cbebae0345626e89",
    "42b9652e488347d1247e4ea4dba90e0434c5d77f",
    "30cf52ecd6f3afb7d6d227dd305c0ecb0d795338",
    "a22251d1e53cd408d3d71037d5fc2d7e20f7c04f",
    "aa79aacca69f63c48e3b39b558578fbcceade8c7",
    "9be8d15a71c2b50f877f222c143710e04d500b50",
    "6877c55a906e183e2c1b13b681c6595a38003e73",
    "57092b3db3a1c4fa0f5d7af603df3c0760d9970e",
    "30708dc26c216b9952a5b6df340d8929af2acd3a",
    "b68ba0c2abdc055ae83b827c2b37196c1a846c8a",
    "7e7f80e83a2ae342a936b2e5d0fd4e9603839893",
    "a29197f9d99cd1d17e7078b1fa05f3af068131d3",
    "d6eccef7d3df7220ba91723d61628cb6139b3f90",
    "27a3a750f76b331c409f94756ebb05c5f7c2b931",
    "d1d876d1e126f582c3def34208b315ad5edb4c88",
    "d4ab103f7e02e3775b623788f6f9e5ba9b7e7971",
    "ab46b898259dd8c24af82b458a4e7789ecbdd5a8",
    "420b4d05562efb6837c840177372a0557d88896d",
    "ec10be885226ad69cb159fc07b90e0a45e23aba3",
    "16ec7dd25eb7943c6ecd4e49ffcac174ae8b7585",
    "f8d9b0a7b274f13c9a1ae04cac9faa2a9f249a44",
    "b9bba499ecff849216b0dc26f97b7692d2fae527",
    "69a41173e7ea924666119ddd34f529e1dc729931",
    "925b11aedb067767a4143ef11acba537f488df38",
    "f717fe829d7618ddd937593a90ea3d4f2ab0887d",
    "c7e783b0957697d36418cbcbe09e9f28d2b2404d",
    "67b43fa5eff3444fe134fef240c0144f098b4fe9",
    "9a354fd185d78ff7b328fe1f24ef59f037981703",
    "ce68f1e68bd894d8d7fa60225a2d91d1d619cc58",
    "bda89330dcf8f94189657a9beb11205cedf0ee21",
    "8dca3d400e6125cb0f4650320e4dbb5b7d5a742c",
    "9ef4ae694a1c16570f48b848b43041d9e2377200",
    "d473435e16bac9dd199a3bbb3091e99b32097944",
    "67be8e78f9050faa383b2f71ec7886822b46cdee",
    "de2d26ec247622769033a4709cb28e63d3d1edd9",
    "bf8e3569fe36b8dd94ce5b1f739d5fcef8e6969c",
    "5d0500f80c6ffb66e200c407479d6be53a31e0af",
    "49faf2eca935ccef1101b158a4c0a7a10eb5eafd",
    "3e6bae8a176c77c6ba66621e1d342b83155ac87d",
    "8fcb4c40ca1bf97c6611633c84561874880d0bde",
    "0e8217209164559d096d6217d41b7d14a61e3977",
    "024640b905e4edb760877a2428f2600e2ef861f3",
    "c7c1a76ca4303b84b1fcdaf9b918a3211d00e0cd",
    "ff099e8638edadf3b216bb68c7359f5d4e533ea2",
    "18b8d51441be6f7a2fca1c56e52153e897ebd96a",
    "a9398b03040011aa92e9a8142d58eef6a31b4db6",
    "84775c51b51cb72611687494f9900be40ea99c00",
    "bbfb49319398e906cc1f81e204c37c68f6ac6c80",
    "8800cbd9c2c64ca69e261779279143eff441715c",
    "74695190791b2bbfa947c8e4c8c1223777e0cc78",
    "6fb0b275e3b175072f0e44049736287dde6ad6ab",
    "ed98ae6581973c7e4110a93b2d7320b780d6b21e",
    "57e6ee0510cc664319df8b5e7bb284294a9d1e46",
    "0a84593f0e676bbf2f4465573438bc9c188e6469",
    "e023542cc6eb2c83f94d44b07b5a2dd502068b2d",
    "12c3a45c36fde0b1a3850086a915dc258379747c",
    "64843eb88653d116e5f52d7ded1bf88a28744e0c",
    "44366cec7f240dd32de1b83f8e8415566b3c1506",
    "af810939689ee15d6b281b7fcff8e1e717900691",
    "2c6c7097105e6bc2395b310e7599923f166e1369",
    "b193335a06f81764c4c08d30ce2e13e4ef32cc23",
    "d788fec4b82abaf068f45bd7e2b256e03fb8c564",
    "0cf730e6f56ad8992a781b983b719048825be68f",
    "0ed073e9a00737cf44561c7ec447d0283dbf1ef1",
    "56dd07ee442ab9f5bed460c131a87b6f1919f229",
    "ae5dc61611c0fe3bc0104b9da2fd9afa1e87ecb5",
    "2c941c441eaa08f5e0eadb66abe2fdee0fb33c59",
    "3c82c182b9c207a6186f4853d67b016b2b0d051f",
    "38bc8b7cc7540e9b131de65c5771fc662feb800b",
    "4ea5997d7eb320483072c57019b4d40ae2bd4c11",
    "697a503f4a99545564987c5c4ce6c48d9ef836f7",
    "b6527c725ed6fffaa9f050e6e1f9a42dc2ad6fac",
    "037561b2b27c460e2ffe1f9ce88510d03272fa50",
    "cb6ba17ffb24818aded80f59f5bf3a120766b137",
    "64c7cbfdae02a9a47eda84ae7a467d336f177c8e",
    "4b4eed95bb6e4ed3b59f2556aaf22494734ea34b",
    "8f9e159a388602d1cd8827203bb1efa6a888dc23",
    "00f9b75a8bb04bfd5d2294ec7e64363998fe6241",
    "67e1016a3b4a618310accb9c015fa8ccefd65e4a",
]


def modify(commits: list[dict[str, object]]) -> None:
    # Drop the bad commits in place, so the caller's list reflects the removal.
    commits[:] = [
        metadata for metadata in commits if metadata["commit"] not in BAD_COMMITS
    ]
    for metadata in commits:
        metadata["authorDate"] = to_san_francisco_time(str(metadata["authorDate"]))
        metadata["committerDate"] = to_san_francisco_time(
            str(metadata["committerDate"])
        )


def generate_rebase_todo(commits: list[dict[str, object]]) -> str:
    lines = []
    for metadata in commits:
        lines.append(f"pick {metadata['commit']}")

        if not metadata["signedOffBy"]:
            committer = " ".join(
                [
                    f"GIT_COMMITTER_NAME={shlex.quote(str(metadata['authorName']))}",
                    f"GIT_COMMITTER_EMAIL={shlex.quote(str(metadata['authorEmail']))}",
                ]
            )
            lines.append(
                f"exec {committer} git commit --amend --no-edit --allow-empty --signoff"
            )

        env = " ".join(
            [
                f"GIT_AUTHOR_NAME={shlex.quote(str(metadata['authorName']))}",
                f"GIT_AUTHOR_EMAIL={shlex.quote(str(metadata['authorEmail']))}",
                f"GIT_AUTHOR_DATE={shlex.quote(str(metadata['authorDate']))}",
                f"GIT_COMMITTER_NAME={shlex.quote(str(metadata['committerName']))}",
                f"GIT_COMMITTER_EMAIL={shlex.quote(str(metadata['committerEmail']))}",
                f"GIT_COMMITTER_DATE={shlex.quote(str(metadata['committerDate']))}",
            ]
        )
        # --reset-author is required: git commit --amend otherwise keeps the
        # picked commit's author and ignores the GIT_AUTHOR_* env vars.
        lines.append(f"exec {env} git commit --amend --no-edit --reset-author")
    return "\n".join(lines) + "\n"


def generate_json(commits: list[dict[str, object]]) -> str:
    return json.dumps({"commits": commits}, indent=2) + "\n"


def main() -> None:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("branch", help="branch to export commits from (e.g. main)")
    parser.add_argument("output", help="path to write the history to")
    parser.add_argument(
        "-C", "--worktree", default=".", help="path to the git worktree"
    )
    parser.add_argument(
        "--format",
        choices=["rebase-todo", "json"],
        default="rebase-todo",
        help="output format (default: rebase-todo)",
    )
    args = parser.parse_args()

    commits = read_commits(args.branch, worktree=args.worktree)

    modify(commits)

    if args.format == "json":
        output = generate_json(commits)
    else:
        output = generate_rebase_todo(commits)
    with open(args.output, "w") as f:
        f.write(output)


if __name__ == "__main__":
    main()
