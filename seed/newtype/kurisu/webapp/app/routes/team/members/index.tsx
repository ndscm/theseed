import React, { useCallback, useEffect } from "react"
import { useTranslation } from "react-i18next"
import { useNavigate } from "react-router"

import tw from "../../../../../../../devprod/ts/grouping-tailwind"
import { useHooinRosterService } from "../../../../../../hooin/roster/client/tsx/HooinRosterServiceContext"
import { type TeamMember } from "../../../../../../hooin/roster/proto/roster_pb"
import KurisuAvatar from "../../../../components/KurisuAvatar"
import OrganicUtils from "../../../../utils/OrganicUtils"

const TeamMembersPage: React.FC<{}> = ({}) => {
  const { t } = useTranslation("team")
  const navigate = useNavigate()
  const rosterService = useHooinRosterService()
  const [teamMembers, setTeamMembers] = React.useState<TeamMember[]>([])

  const reload = useCallback(async () => {
    if (!rosterService) {
      return
    }
    const listResponse = await rosterService.ListTeamMembers()
    setTeamMembers(listResponse.teamMembers)
  }, [rosterService])

  useEffect(() => {
    reload()
  }, [reload])

  return (
    <main
      className={tw({
        layout: "min-h-0 flex-1 overflow-y-auto",
      })}
    >
      <table
        className={tw({
          component: "table-pin-rows table",
        })}
      >
        <thead>
          <tr>
            <th
              className={tw({
                appearance:
                  "text-neutral text-xs font-semibold tracking-wider uppercase",
              })}
            >
              {t("member.nameLabel", "Name")}
            </th>
            <th
              className={tw({
                appearance:
                  "text-neutral text-xs font-semibold tracking-wider uppercase",
              })}
            >
              {t("member.statusLabel", "Status")}
            </th>
          </tr>
        </thead>
        <tbody>
          {teamMembers.map((member) => (
            <tr
              key={member.personId}
              onClick={() => navigate(`/person/${member.handle}`)}
              className={tw({
                appearance: "bg-base-100 cursor-pointer",
                state: "hover:bg-base-200",
              })}
            >
              <td>
                <div
                  className={tw({
                    layout: "flex items-center gap-3",
                  })}
                >
                  <KurisuAvatar
                    size="small"
                    organic={OrganicUtils.GetOrganic(member.organic)}
                  >
                    <span
                      className={tw({
                        appearance: "text-sm font-semibold",
                      })}
                    >
                      {member.displayName?.charAt(0)?.toUpperCase() ?? "?"}
                    </span>
                  </KurisuAvatar>
                  <div className={tw({ layout: "min-w-0" })}>
                    <div
                      className={tw({
                        appearance: "text-base-content text-sm font-semibold",
                      })}
                    >
                      {member.displayName}
                    </div>
                    <div
                      className={tw({
                        appearance: "text-neutral font-mono text-xs",
                      })}
                    >
                      {member.handle}@
                    </div>
                  </div>
                </div>
              </td>
              <td>
                {member.onDuty ? (
                  <span
                    className={tw({
                      component: "badge badge-success badge-sm",
                      layout: "gap-1.5",
                    })}
                  >
                    {t("member.onDutyStatus", "On Duty")}
                  </span>
                ) : (
                  <span
                    className={tw({
                      component: "badge badge-ghost badge-sm",
                      layout: "gap-1.5",
                    })}
                  >
                    {t("member.offDutyStatus", "Off Duty")}
                  </span>
                )}
              </td>
            </tr>
          ))}
        </tbody>
      </table>
      {teamMembers.length === 0 && (
        <div
          className={tw({
            layout: "flex flex-col items-center justify-center py-16",
          })}
        >
          <div
            className={tw({
              layout: "mb-1",
              appearance: "text-base font-semibold",
            })}
          >
            {t("members.noMembersHint", "No members")}
          </div>
        </div>
      )}
    </main>
  )
}

export default TeamMembersPage
