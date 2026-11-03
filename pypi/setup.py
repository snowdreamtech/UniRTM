from setuptools import setup

try:
    from wheel.bdist_wheel import bdist_wheel as _bdist_wheel

    class bdist_wheel(_bdist_wheel):  # noqa: N801
        def finalize_options(self):
            super().finalize_options()
            self.root_is_pure = False

        def get_tag(self):
            python, abi, plat = super().get_tag()
            return "py3", "none", plat
except ImportError:
    bdist_wheel = None

cmdclass = {}
if bdist_wheel:
    cmdclass["bdist_wheel"] = bdist_wheel

setup(
    packages=["snowdreamtech_unirtm"],
    package_data={"snowdreamtech_unirtm": ["bin/*"]},
    cmdclass=cmdclass,
)
