import React, { useCallback, useEffect } from "react"
import { useTranslation } from "react-i18next"
import { NavLink } from "react-router"

import { PlusIcon, UsersIcon } from "lucide-react"

import tw from "../../../../devprod/ts/grouping-tailwind/dist"
import { useHooinRosterService } from "../../../hooin/roster/client/tsx/HooinRosterServiceContext"
import { type Team } from "../../../hooin/roster/proto/roster_pb"
import KurisuAvatar from "./KurisuAvatar"
import KurisuTopBar from "./KurisuTopBar"

const TeamHeader: React.FC = () => {
  const { t } = useTranslation("team")
  const rosterService = useHooinRosterService()
  const [team, setTeam] = React.useState<Team | null>(null)

  const reload = useCallback(async () => {
    if (!rosterService) {
      return
    }
    const tmpTeam = await rosterService.GetTeam()
    setTeam(tmpTeam)
  }, [rosterService])

  useEffect(() => {
    reload()
  }, [reload])

  return (
    <header className={tw({ layout: "sticky top-0" })}>
      <KurisuTopBar
        title={
          <>
            {t("breadcrumb.team", "Team")}
            {team && (
              <span
                className={tw({
                  layout: "ml-2",
                  appearance: "text-neutral text-base font-normal",
                })}
              >
                @{team.handle}
              </span>
            )}
          </>
        }
      />
      <div
        className={tw({
          layout: "shrink-0 px-7 pt-5",
          appearance: "bg-base-100",
          state: "max-sm:px-5 max-sm:pt-4",
        })}
      >
        <div
          className={tw({
            layout: "mb-5 flex flex-wrap items-center gap-5",
          })}
        >
          <KurisuAvatar size="large">
            <UsersIcon className={tw({ layout: "size-8" })} />
          </KurisuAvatar>
          <div className={tw({ layout: "min-w-60 flex-1" })}>
            <div
              className={tw({
                layout: "flex flex-wrap items-baseline gap-2",
              })}
            >
              <h2
                className={tw({
                  layout: "m-0",
                  appearance:
                    "text-base-content text-2xl font-semibold tracking-tight",
                })}
              >
                {team?.displayName ?? t("info.displayNamePlaceholder", "Team")}
              </h2>
              {team && (
                <span
                  className={tw({
                    appearance: "text-neutral font-mono text-sm font-medium",
                  })}
                >
                  @{team.handle}
                </span>
              )}
            </div>
          </div>
          <div className={tw({ layout: "ml-auto flex items-center gap-2" })}>
            <button className={tw({ component: "btn btn-primary btn-sm" })}>
              <PlusIcon className={tw({ layout: "size-4" })} />
              {t("members.addMemberButton", "Add member")}
            </button>
          </div>
        </div>
        <div role="tablist" className={tw({ component: "tabs tabs-border" })}>
          <NavLink
            to="/team/members"
            role="tab"
            className={({ isActive }) =>
              tw({ component: "tab" }, isActive && { component: "tab-active" })
            }
          >
            {t("members.title", "Members")}
          </NavLink>
        </div>
      </div>
    </header>
  )
}

export default TeamHeader
