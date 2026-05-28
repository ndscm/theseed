import React, { useCallback, useEffect } from "react"
import { useTranslation } from "react-i18next"
import { Link, NavLink, type NavLinkProps, useNavigate } from "react-router"

import {
  HouseIcon,
  MoonIcon,
  PanelLeftCloseIcon,
  PanelLeftOpenIcon,
  SunIcon,
  SunMoonIcon,
  UsersIcon,
} from "lucide-react"

import LoginButton from "../../../../cloud/login/client/tsx/LoginButton"
import { useLoginService } from "../../../../cloud/login/client/tsx/LoginServiceContext"
import { type LoginStatus } from "../../../../cloud/login/proto/login_pb"
import tw from "../../../../devprod/ts/grouping-tailwind"
import { useHooinRosterService } from "../../../hooin/roster/client/tsx/HooinRosterServiceContext"
import { type TeamMember } from "../../../hooin/roster/proto/roster_pb"
import KurisuAvatar from "./KurisuAvatar"

type ColorScheme = "" | "light" | "dark"

const SideAppRow: React.FC<{}> = ({}) => {
  const { t } = useTranslation("common")
  const [colorScheme, setColorScheme] = React.useState<ColorScheme>("")

  const systemIsDark =
    typeof window !== "undefined" &&
    window.matchMedia("(prefers-color-scheme: dark)").matches

  const toggleTheme = () => {
    const rotation: { [key in ColorScheme]: ColorScheme } = !systemIsDark
      ? {
          "": "dark",
          dark: "light",
          light: "",
        }
      : {
          "": "light",
          light: "dark",
          dark: "",
        }
    const next = rotation[colorScheme]
    switch (next) {
      case "":
        document.documentElement.removeAttribute("data-theme")
        break
      case "light":
        document.documentElement.setAttribute("data-theme", "kurisu-light")
        break
      case "dark":
        document.documentElement.setAttribute("data-theme", "kurisu-dark")
        break
    }
    setColorScheme(next)
  }

  const ColorSchemeIcon =
    colorScheme === "light"
      ? SunIcon
      : colorScheme === "dark"
        ? MoonIcon
        : SunMoonIcon

  return (
    <div
      className={tw({
        layout: "flex items-center pb-6",
      })}
    >
      <div
        className={tw({
          layout: "flex w-full items-center gap-2",
        })}
      >
        <Link
          to="/"
          className={tw({
            layout: "flex items-center gap-2 px-2",
            appearance: "no-underline transition-opacity duration-200",
            state: "is-drawer-close:hidden",
          })}
        >
          <KurisuAvatar size="small">
            <span className={tw({ appearance: "text-base font-bold" })}>K</span>
          </KurisuAvatar>
          <span
            className={tw({
              appearance: "text-base-content text-lg font-bold tracking-tight",
            })}
          >
            {t("system.brand", "Kurisu")}
          </span>
        </Link>
      </div>
      <div>
        <div className={tw({ layout: "ml-auto flex items-center gap-1" })}>
          <button
            type="button"
            className={tw({
              layout: "cursor-pointer rounded-lg border-none p-3",
              appearance: "text-neutral-content bg-transparent",
              state:
                "hover:bg-base-200 hover:text-base-content is-drawer-close:hidden",
            })}
            onClick={toggleTheme}
          >
            <ColorSchemeIcon className={tw({ layout: "h-5 w-5 shrink-0" })} />
          </button>
          <label
            htmlFor="kurisu-sidebar-toggle"
            className={tw({
              layout: "cursor-pointer rounded-lg p-3",
              appearance: "text-neutral-content",
              state:
                "hover:bg-base-200 hover:text-base-content is-drawer-close:hidden",
            })}
          >
            <PanelLeftCloseIcon
              className={tw({ layout: "h-5 w-5 shrink-0" })}
            />
          </label>
          <label
            htmlFor="kurisu-sidebar-toggle"
            className={tw({
              layout: "hidden cursor-pointer rounded-lg p-3",
              appearance: "text-neutral-content",
              state:
                "is-drawer-close:flex hover:bg-base-200 hover:text-base-content",
            })}
          >
            <PanelLeftOpenIcon className={tw({ layout: "h-5 w-5 shrink-0" })} />
          </label>
        </div>
      </div>
    </div>
  )
}

const SideNavLink: React.FC<
  { icon: React.ReactNode } & NavLinkProps &
    React.AnchorHTMLAttributes<HTMLAnchorElement>
> = ({ icon, to, end, children }) => (
  <li>
    <NavLink
      to={to}
      end={end}
      className={({ isActive }) =>
        tw(
          {
            layout: "flex items-center gap-3 p-3",
            appearance:
              "rounded-lg text-sm font-medium no-underline transition-colors",
          },
          isActive
            ? { appearance: "bg-accent text-accent-content font-semibold" }
            : {
                appearance: "text-neutral-content",
                state: "hover:bg-base-200 hover:text-base-content",
              },
        )
      }
    >
      {icon}
      <span
        className={tw({
          appearance: "transition-opacity duration-200",
          state: "is-drawer-close:opacity-0",
        })}
      >
        {children}
      </span>
    </NavLink>
  </li>
)

const SidePeopleList: React.FC = () => {
  const { t } = useTranslation("common")
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
    <div
      className={tw({
        layout: "flex min-h-0 flex-1 flex-col",
      })}
    >
      <div
        className={tw({
          layout: "shrink-0 px-3 pb-2",
          appearance:
            "text-neutral text-xs font-semibold tracking-wider uppercase",
          state: "is-drawer-close:opacity-0",
        })}
      >
        {t("side.people", "People")}
      </div>
      <div
        className={tw({
          layout:
            "flex min-h-0 flex-1 flex-col gap-1 overflow-x-hidden overflow-y-auto",
        })}
      >
        {teamMembers.map((member) => (
          <NavLink
            key={member.personId}
            to={`/person/${member.handle}`}
            className={({ isActive }) =>
              tw(
                {
                  layout: "flex items-center gap-2 rounded-lg p-1",
                  appearance: "no-underline",
                },
                isActive
                  ? {
                      appearance: "bg-accent text-accent-content font-semibold",
                    }
                  : {
                      appearance: "text-neutral-content",
                      state: "hover:bg-base-200 hover:text-base-content",
                    },
              )
            }
          >
            <KurisuAvatar size="small" organic="silicon">
              <span
                className={tw({
                  appearance: "text-base font-bold",
                })}
              >
                {member.displayName?.charAt(0)?.toUpperCase() ?? "?"}
              </span>
            </KurisuAvatar>
            <span
              className={tw({
                layout: "flex min-w-0 flex-1 flex-col",
                appearance: "leading-tight",
                state: "is-drawer-close:hidden",
              })}
            >
              <span
                className={tw({
                  appearance:
                    "text-base-content truncate text-sm font-semibold",
                })}
              >
                {member.displayName}
              </span>
              <span
                className={tw({
                  appearance: "text-neutral truncate font-mono text-xs",
                })}
              >
                {member.handle}@
              </span>
            </span>
            <span
              className={tw({
                layout: "m-3 size-2 shrink-0 rounded-full",
                appearance: member.onDuty ? "bg-success" : "bg-base-300",
                state: "is-drawer-close:hidden",
              })}
            />
          </NavLink>
        ))}
      </div>
    </div>
  )
}

const SideLoginButton: React.FC<{}> = ({}) => {
  const { t } = useTranslation("common")
  const navigate = useNavigate()
  const loginService = useLoginService()
  const [loginStatus, setLoginStatus] = React.useState<LoginStatus>()

  useEffect(() => {
    void (async () => {
      if (!loginService) {
        return
      }
      const currentLoginStatus = await loginService.GetLoginStatus()
      setLoginStatus(currentLoginStatus)
    })()
  }, [loginService])

  return (
    <LoginButton
      className={tw({
        state: "hover:bg-base-200",
        layout: "flex w-full cursor-pointer items-center gap-3",
        appearance: "rounded-lg transition-colors",
      })}
      loading={
        <span
          className={tw({
            component: "loading loading-spinner loading-sm",
          })}
        />
      }
      anonymous={t("auth.loginButton", "Login")}
      onClick={() => {
        navigate(`/person/${loginStatus?.userHandle}`)
      }}
    >
      <KurisuAvatar size="small" organic="carbon">
        <span className={tw({ appearance: "text-base font-bold" })}>
          {loginStatus?.displayName?.charAt(0)?.toUpperCase() ?? "?"}
        </span>
      </KurisuAvatar>
      <div
        className={tw({
          state: "is-drawer-close:hidden",
          layout: "min-w-0 flex-1",
        })}
      >
        <div
          className={tw({
            appearance: "text-base-content truncate text-sm font-semibold",
          })}
        >
          {loginStatus?.displayName}
        </div>
        <div
          className={tw({
            appearance: "text-neutral truncate text-xs",
          })}
        >
          {loginStatus?.userHandle}@
        </div>
      </div>
    </LoginButton>
  )
}

const KurisuSideBar: React.FC = () => {
  const { t } = useTranslation("common")

  return (
    <nav
      className={tw({
        layout: "flex min-h-full flex-col px-3 py-3",
        appearance: "bg-base-100 border-base-300 border-r",
        state: "is-drawer-close:w-17 is-drawer-open:w-60",
      })}
    >
      <SideAppRow />
      <ul
        className={tw({
          layout: "m-0 flex w-full shrink-0 list-none flex-col gap-1 p-0",
        })}
      >
        <SideNavLink
          to="/"
          end
          icon={<HouseIcon className={tw({ layout: "h-5 w-5 shrink-0" })} />}
        >
          {t("page.home", "Home")}
        </SideNavLink>
        <SideNavLink
          to="/team"
          icon={<UsersIcon className={tw({ layout: "h-5 w-5 shrink-0" })} />}
        >
          {t("page.team", "Team")}
        </SideNavLink>
      </ul>
      <div
        className={tw({
          layout: "mx-1 my-2 shrink-0",
          appearance: "bg-base-300 h-px",
          state: "is-drawer-close:opacity-0",
        })}
      />
      <SidePeopleList />
      <div
        className={tw({
          layout: "mx-1 my-2 shrink-0",
          appearance: "bg-base-300 h-px",
          state: "is-drawer-close:opacity-0",
        })}
      />
      <div
        className={tw({
          layout: "my-1 shrink-0 p-1",
        })}
      >
        <SideLoginButton />
      </div>
    </nav>
  )
}

export default KurisuSideBar
