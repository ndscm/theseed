"""Rules for declaring and composing service stacks.

A stack is a set of services, their server providers, and client dependency
wiring. seed_stack_service and seed_stack_client declare individual pieces;
seed_stack composes them into a merged view with optional overrides.
"""

SeedStackServiceInfo = provider(
    "Metadata for a single service declaration.",
    fields = {
        "label": "string label of this service target",
        "depends": "dict of service label string -> client config",
    },
)

SeedStackClientInfo = provider(
    "A client's dependency declarations on one or more services, including connection and auth config.",
    fields = {
        "label": "string label of this client target",
        "depend_services": "dict of service label string -> connection config",
    },
)

SeedStackServiceProviderInfo = provider(
    "A server or container that provides one or more services.",
    fields = {
        "label": "string label of this provider target",
        "provider_type": "string, e.g. 'server' or 'container'",
        "connect": "dict of protocol -> config",
        "provides": "list of service label strings this provider serves",
    },
)

SeedStackInfo = provider(
    "A composed view of services, their providers, and any overrides.",
    fields = {
        "services": "dict of service label string -> struct(providers, depends)",
        "providers": "dict of provider label string -> struct(provider_type, connect)",
        "overrides": "dict of service label string -> override config",
    },
)

def _normalize_label(label):
    s = str(label)
    repo_end = s.find("//")
    if repo_end > 0:
        s = s[repo_end:]
    return s

def _merge_stack_infos(infos):
    services = {}
    providers = {}
    overrides = {}
    for info in infos:
        services.update(info.services)
        providers.update(info.providers)
        overrides.update(info.overrides)
    return SeedStackInfo(
        services = services,
        providers = providers,
        overrides = overrides,
    )

def _stack_info_to_json(info):
    services = {}
    for label, svc in info.services.items():
        services[label] = {
            "providers": svc.providers,
            "depends": svc.depends,
        }
    providers = {}
    for label, p in info.providers.items():
        providers[label] = {
            "provider_type": p.provider_type,
            "connect": p.connect,
        }
    result = {
        "services": services,
        "providers": providers,
    }
    if info.overrides:
        result["overrides"] = info.overrides
    return json.encode_indent(result)

# --- seed_stack_service ---

def _seed_stack_service_impl(ctx):
    label = _normalize_label(ctx.label)
    depends = {}
    transitive_infos = []
    for dep in ctx.attr.depend_clients:
        ci = dep[SeedStackClientInfo]
        depends.update(ci.depend_services)
        if SeedStackInfo in dep:
            transitive_infos.append(dep[SeedStackInfo])
    service_info = SeedStackServiceInfo(
        label = label,
        depends = depends,
    )
    result = [service_info]
    if transitive_infos:
        result.append(_merge_stack_infos(transitive_infos))
    return result

_seed_stack_service_rule = rule(
    implementation = _seed_stack_service_impl,
    attrs = {
        "depend_clients": attr.label_list(providers = [SeedStackClientInfo]),
    },
)

def seed_stack_service(
        name,
        depend_clients = {},
        **kwargs):
    _seed_stack_service_rule(
        name = name,
        depend_clients = depend_clients.keys() if type(depend_clients) == type({}) else depend_clients,
        **kwargs
    )

# --- seed_stack_service_provider ---

def _seed_stack_service_provider_impl(ctx):
    label = _normalize_label(ctx.label)
    provides = [_normalize_label(p.label) for p in ctx.attr.provides]
    return [SeedStackServiceProviderInfo(
        label = label,
        provider_type = ctx.attr.provider_type,
        connect = json.decode(ctx.attr.connect_json),
        provides = provides,
    )]

_seed_stack_service_provider_rule = rule(
    implementation = _seed_stack_service_provider_impl,
    attrs = {
        "provider_type": attr.string(default = "server"),
        "connect_json": attr.string(),
        "provides": attr.label_list(providers = [SeedStackServiceInfo]),
    },
)

def seed_stack_service_provider(
        name,
        provider_type = "server",
        connect = {},
        provides = [],
        **kwargs):
    _seed_stack_service_provider_rule(
        name = name,
        provider_type = provider_type,
        connect_json = json.encode(connect),
        provides = provides,
        **kwargs
    )

# --- seed_stack_client ---

def _seed_stack_client_impl(ctx):
    transitive_infos = []
    for dep in ctx.attr.stacks:
        transitive_infos.append(dep[SeedStackInfo])
    result = [SeedStackClientInfo(
        label = str(ctx.label),
        depend_services = json.decode(ctx.attr.depend_services_json),
    )]
    if transitive_infos:
        result.append(_merge_stack_infos(transitive_infos))
    return result

_seed_stack_client_rule = rule(
    implementation = _seed_stack_client_impl,
    attrs = {
        "depend_services_json": attr.string(),
        "stacks": attr.label_list(providers = [SeedStackInfo]),
    },
)

def seed_stack_client(
        name,
        depend_services = {},
        **kwargs):
    stack_labels = []
    for svc_config in depend_services.values():
        if type(svc_config) == type({}) and "stack" in svc_config:
            stack_labels.append(svc_config["stack"])
    _seed_stack_client_rule(
        name = name,
        depend_services_json = json.encode(depend_services),
        stacks = stack_labels,
        **kwargs
    )

# --- seed_stack ---

def _seed_stack_impl(ctx):
    transitive_infos = []
    for dep in ctx.attr.stacks:
        transitive_infos.append(dep[SeedStackInfo])
    for dep in ctx.attr.services:
        if SeedStackInfo in dep:
            transitive_infos.append(dep[SeedStackInfo])

    services = {}
    providers = {}
    overrides = {}
    for info in transitive_infos:
        services.update(info.services)
        providers.update(info.providers)
        overrides.update(info.overrides)

    provider_to_services = {}
    for dep in ctx.attr.providers:
        pi = dep[SeedStackServiceProviderInfo]
        providers[pi.label] = struct(
            provider_type = pi.provider_type,
            connect = pi.connect,
        )
        for svc_label in pi.provides:
            if svc_label not in provider_to_services:
                provider_to_services[svc_label] = []
            provider_to_services[svc_label].append(pi.label)

    for dep in ctx.attr.services:
        si = dep[SeedStackServiceInfo]
        services[si.label] = struct(
            providers = provider_to_services.get(si.label, []),
            depends = si.depends,
        )

    overrides.update(json.decode(ctx.attr.overrides_json))

    stack_info = SeedStackInfo(
        services = services,
        providers = providers,
        overrides = overrides,
    )
    out = ctx.actions.declare_file(ctx.label.name + ".json")
    ctx.actions.write(out, _stack_info_to_json(stack_info))
    return [DefaultInfo(files = depset([out])), stack_info]

_seed_stack_rule = rule(
    implementation = _seed_stack_impl,
    attrs = {
        "services": attr.label_list(providers = [SeedStackServiceInfo]),
        "providers": attr.label_list(providers = [SeedStackServiceProviderInfo]),
        "stacks": attr.label_list(providers = [SeedStackInfo]),
        "overrides_json": attr.string(),
    },
)

def seed_stack(
        name,
        services = [],
        providers = [],
        stacks = [],
        overrides = {},
        **kwargs):
    _seed_stack_rule(
        name = name,
        services = services,
        providers = providers,
        stacks = stacks,
        overrides_json = json.encode(overrides),
        **kwargs
    )
