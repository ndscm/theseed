import React, { useCallback, useEffect } from "react"
import { useTranslation } from "react-i18next"
import { NavLink, useParams } from "react-router"

import { ChevronRightIcon } from "lucide-react"

import tw from "../../../../devprod/ts/grouping-tailwind/dist"
import { useHooinRosterService } from "../../../hooin/roster/client/tsx/HooinRosterServiceContext"
import { type TeamMember } from "../../../hooin/roster/proto/roster_pb"
import OrganicUtils from "../utils/OrganicUtils"
import KurisuAvatar from "./KurisuAvatar"
import KurisuTopBar from "./KurisuTopBar"

const PersonHeader: React.FC = () => {
  const { t } = useTranslation("person")
  const { handle } = useParams<{ handle: string }>()
  const rosterService = useHooinRosterService()
  const [person, setPerson] = React.useState<TeamMember | null>(null)

  const reload = useCallback(async () => {
    setPerson(null)
    if (!rosterService || !handle) {
      return
    }
    const listResponse = await rosterService.ListTeamMembers()
    const found = listResponse.teamMembers.find((m) => m.handle === handle)
    if (found) {
      setPerson(found)
    }
  }, [rosterService, handle])

  useEffect(() => {
    reload()
  }, [reload])

  return (
    <header className={tw({ layout: "sticky top-0" })}>
      <KurisuTopBar
        title={
          <div
            className={tw({
              layout: "flex items-center gap-2",
              appearance: "whitespace-nowrap",
            })}
          >
            <NavLink
              to="/person"
              className={tw({
                appearance: "text-neutral text-base font-semibold no-underline",
                state: "hover:text-base-content",
              })}
            >
              {t("breadcrumb.person", "Person")}
            </NavLink>
            <ChevronRightIcon
              className={tw({
                layout: "size-3.5",
                appearance: "text-neutral",
              })}
            />
            <span
              className={tw({
                appearance: "text-base-content",
              })}
            >
              {person?.displayName ?? handle}
            </span>
          </div>
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
          <KurisuAvatar
            size="large"
            organic={OrganicUtils.GetOrganic(person?.organic)}
          >
            <span
              className={tw({
                appearance: "text-xl font-semibold",
              })}
            >
              {person?.displayName?.charAt(0)?.toUpperCase() ??
                handle?.charAt(0)?.toUpperCase() ??
                "?"}
            </span>
          </KurisuAvatar>
          <div className={tw({ layout: "min-w-60 flex-1" })}>
            <div
              className={tw({
                layout: "flex flex-wrap items-center gap-3",
              })}
            >
              <h2
                className={tw({
                  layout: "m-0",
                  appearance:
                    "text-base-content text-2xl font-semibold tracking-tight",
                })}
              >
                {person?.displayName ?? handle}
              </h2>
              {person &&
                (person.onDuty ? (
                  <span
                    className={tw({
                      component: "badge badge-success badge-sm",
                      layout: "gap-2",
                    })}
                  >
                    {t("status.onDuty", "On Duty")}
                  </span>
                ) : (
                  <span
                    className={tw({
                      component: "badge badge-ghost badge-sm",
                      layout: "gap-2",
                    })}
                  >
                    {t("status.offDuty", "Off Duty")}
                  </span>
                ))}
            </div>
            <div
              className={tw({
                layout: "mt-1",
                appearance: "text-neutral text-sm",
              })}
            >
              <span className={tw({ appearance: "font-mono" })}>{handle}@</span>
            </div>
          </div>
        </div>
        <div role="tablist" className={tw({ component: "tabs tabs-border" })}>
          <NavLink
            to={`/person/${handle}/chat`}
            role="tab"
            className={({ isActive }) =>
              tw({ component: "tab" }, isActive && { component: "tab-active" })
            }
          >
            {t("chat.title", "Chat")}
          </NavLink>
          <NavLink
            to={`/person/${handle}/brain`}
            role="tab"
            className={({ isActive }) =>
              tw({ component: "tab" }, isActive && { component: "tab-active" })
            }
          >
            {t("brain.title", "Brain")}
          </NavLink>
          <NavLink
            to={`/person/${handle}/memory`}
            role="tab"
            className={({ isActive }) =>
              tw({ component: "tab" }, isActive && { component: "tab-active" })
            }
          >
            {t("memory.title", "Memory")}
          </NavLink>
          <NavLink
            to={`/person/${handle}/workstation`}
            role="tab"
            className={({ isActive }) =>
              tw({ component: "tab" }, isActive && { component: "tab-active" })
            }
          >
            {t("workstation.title", "Workstation")}
          </NavLink>
          <NavLink
            to={`/person/${handle}/terminal`}
            role="tab"
            className={({ isActive }) =>
              tw({ component: "tab" }, isActive && { component: "tab-active" })
            }
          >
            {t("terminal.title", "Terminal")}
          </NavLink>
          <NavLink
            to={`/person/${handle}/attendance`}
            role="tab"
            className={({ isActive }) =>
              tw({ component: "tab" }, isActive && { component: "tab-active" })
            }
          >
            {t("attendance.title", "Attendance")}
          </NavLink>
        </div>
      </div>
    </header>
  )
}

export default PersonHeader
