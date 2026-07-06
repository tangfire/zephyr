type ZephyrLogoProps = {
  className?: string;
  title?: string;
};

export function ZephyrLogo({ className = "", title }: ZephyrLogoProps) {
  return (
    <svg
      className={`zephyr-logo ${className}`.trim()}
      viewBox="0 0 64 64"
      role={title ? "img" : undefined}
      aria-hidden={title ? undefined : true}
      focusable="false"
    >
      {title && <title>{title}</title>}
      <circle className="zephyr-logo-core" cx="32" cy="32" r="27" />
      <path className="zephyr-logo-breeze" d="M17 35c8.2-6.8 21.5-7.2 30-1.1" />
      <path className="zephyr-logo-accent" d="M40 18 28 46" />
    </svg>
  );
}
