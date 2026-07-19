import React, { useCallback, useEffect, useMemo, useState } from "react"
import { useTranslation } from "react-i18next"

import { HouseIcon } from "lucide-react"

import tw from "../../../../../../../../devprod/ts/grouping-tailwind"
import VscodeWebWorkbench from "../../../../../../../../devprod/vscode/web/tsx/VscodeWebWorkbench"
import HooinRaidFileSystem from "../../../../../../../hooin/raid/client/tsx/HooinRaidFileSystem"
import { useHooinRaidService } from "../../../../../../../hooin/raid/client/tsx/HooinRaidServiceContext"
import { useHooinRosterService } from "../../../../../../../hooin/roster/client/tsx/HooinRosterServiceContext"

const PersonWorkstationPage: React.FC<{ params: { handle: string } }> = ({
  params,
}) => {
  const { handle } = params
  const { t } = useTranslation("person")
  const raidService = useHooinRaidService()
  const rosterService = useHooinRosterService()

  const [personId, setPersonId] = useState<string>("")
  // The folder the editor is open on, and, being empty, whether it is open at
  // all: an editor is a folder somebody asked for, and until somebody has, there
  // is a button here instead. A workbench that failed to open leaves it empty
  // again, and the button is what tries once more.
  const [workspacePath, setWorkspacePath] = useState("")
  const [error, setError] = useState("")

  const personHandle = handle
    .trim()
    .replace(/^@/, "")
    .replace(/@$/, "")
    .toLowerCase()
    .trim()

  useEffect(() => {
    void (async () => {
      if (!rosterService) {
        return
      }
      const member = await rosterService.GetTeamMember("", {
        handle: personHandle,
      })
      setPersonId(member.personId)
    })()
  }, [rosterService, personHandle])

  // The workstation is what the editor opens, and this is what it reads it
  // through: the questions the editor asks about a path arrive here, on the
  // page, and are answered over raid. The workbench itself never sees any of it.
  //
  // The editor names the workstation by the handle it was opened on — that is
  // the authority it shows — but raid names it by the person's id, which is what
  // a role is granted on. The map is how the one becomes the other; until the
  // roster has answered there is no id to give, and the editor is not open yet.
  const webFileSystem = useMemo(() => {
    if (!raidService) {
      return null
    }
    return new HooinRaidFileSystem(raidService, { [personHandle]: personId })
  }, [raidService, personHandle, personId])

  // Where the person's home is is the workstation's answer rather than this
  // page's to guess: it is the folder the editor opens, and asking for it is
  // also how the page finds out whether the person is on duty at all — which is
  // why it is asked when somebody wants the editor, and not before.
  const openHome = useCallback(async () => {
    if (!raidService) {
      return
    }

    setError("")
    try {
      const path = await raidService.GetUserHome(personId)
      setWorkspacePath(path)
    } catch (homeError: unknown) {
      setError(
        homeError instanceof Error ? homeError.message : String(homeError),
      )
    }
  }, [raidService, personId])

  return (
    <main className={tw({ layout: "flex min-h-0 flex-1 flex-col" })}>
      {webFileSystem && workspacePath ? (
        <VscodeWebWorkbench
          className={tw({ layout: "min-h-0 flex-1" })}
          webFileSystem={webFileSystem}
          workspaceAuthority={personHandle}
          workspacePath={workspacePath}
          onError={(workbenchError: unknown) => {
            // The workbench never opened, so the page is back to where it was
            // before it was asked to: the folder is let go of, which is what
            // brings the button back.
            setWorkspacePath("")
            setError(
              workbenchError instanceof Error
                ? workbenchError.message
                : String(workbenchError),
            )
          }}
        />
      ) : (
        <div
          className={tw({
            layout:
              "flex min-h-0 flex-1 flex-col items-center justify-center gap-3",
          })}
        >
          <button
            className={tw({ component: "btn btn-primary" })}
            onClick={openHome}
            disabled={!raidService}
          >
            <HouseIcon />
            {t("workstation.home", "Home")}
          </button>
          {error && (
            <span
              className={tw({
                layout: "max-w-md truncate",
                appearance: "text-base-content/60 text-xs",
              })}
            >
              {error}
            </span>
          )}
        </div>
      )}
    </main>
  )
}

export default PersonWorkstationPage
