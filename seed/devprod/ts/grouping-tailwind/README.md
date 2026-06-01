# Grouping Tailwind

> I still don't understand why tailwind chose to drop the type checking
> mechanism and proper transpiler plugins and throw all the dirty stuff to the
> front-end developers. Using tailwind feels like writing assembly for creating
> a software, writing vanilla javascript for creating a website. which may
> brings some performance but sacrafice developer's life. I'm pretty sure the
> tailwind project is truning the clock back and it will be thown into trash bin
> and replaced by some kind of vite plugin. Since the industry is moving towards
> tailwind for now, it's a problem that cannot be avoided to support tailwind.
>
> -- Nagi

Project `grouping-tailwind` is a simple utility to sacrifice a little bit
performance to protect developer's eyes from tailwind classes. It's inspired by
the `cn` util from `shadcn` with a stronger limitation than `clsx`. It introduce
a fake typed map and simply merge the classnames without any real type checking.
Developers could copy-paste the util implementation if they prefer another
grouping rules.
